package proterrors

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/status"

	"github.com/not-for-prod/proterror/proterror"
	"github.com/not-for-prod/proterror/registry"
)

// AsProtError tries to turn an error into a ProtError if possible.
func AsProtError(err error) (error, bool) {
	st, ok := status.FromError(err)
	if ok {
		for _, detail := range st.Details() {
			if registry.Instance().Has(detail) {
				if val, ok := detail.(error); ok {
					return val, true
				}
			}
		}
	}

	if registry.Instance().Has(err) {
		return err, true
	}

	return err, false
}

// AsStatus tries to turn a ProtError into a domain error if possible.
func AsStatus(err error) *status.Status {
	allowed := registry.Instance().Has(err)
	if allowed {
		var val ProtError
		if errors.As(err, &val) {
			return val.Status()
		}
	}

	return (&proterror.Unknown{}).Status()
}

func PGErrorToProtError(err error) error {
	if err == nil {
		return nil
	}

	// Handle "no rows"
	if errors.Is(err, sql.ErrNoRows) {
		return errors.Join(&proterror.NotFound{}, err)
	}

	// Try to unwrap Postgres derror
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return errors.Join(&proterror.AlreadyExists{}, err)

		// Missing objects / references
		case pgerrcode.ForeignKeyViolation, pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
			return errors.Join(&proterror.NotFound{}, err)

		// Invalid input / bad request
		case pgerrcode.NotNullViolation,
			pgerrcode.CheckViolation,
			pgerrcode.InvalidTextRepresentation,
			pgerrcode.NumericValueOutOfRange,
			pgerrcode.InvalidParameterValue:
			return errors.Join(&proterror.InvalidArgument{}, err)

		// Permission / auth issues
		case pgerrcode.InsufficientPrivilege:
			return errors.Join(&proterror.PermissionDenied{}, err)
		case pgerrcode.InvalidPassword:
			return errors.Join(&proterror.Unauthenticated{}, err)

		// Timeouts / cancellations / deadlocks
		case pgerrcode.DeadlockDetected, pgerrcode.QueryCanceled:
			return errors.Join(&proterror.DeadlineExceeded{}, err)

		// Resource / availability problems
		case pgerrcode.TooManyConnections, pgerrcode.CannotConnectNow:
			return errors.Join(&proterror.Unavailable{}, err)
		}
	}

	return errors.Join(&proterror.Internal{}, err)
}
