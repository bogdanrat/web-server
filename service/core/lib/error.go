package lib

import (
	"github.com/bogdanrat/web-server/contracts/models"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleRPCError(err error) *models.JSONError {
	if err != nil {
		var jsonErr *models.JSONError
		errorCode := status.Code(err)

		switch errorCode {
		case codes.InvalidArgument, codes.ResourceExhausted:
			errorStatus := status.Convert(err)
			for _, details := range errorStatus.Details() {
				switch info := details.(type) {
				case *epb.BadRequest_FieldViolation:
					jsonErr = models.NewBadRequestError(info.Description, info.Field)
				default:
					jsonErr = models.NewInternalServerError(err.Error())
				}
			}
		case codes.PermissionDenied:
			jsonErr = models.NewUnauthorizedError(err.Error())
		case codes.NotFound:
			errorStatus := status.Convert(err)
			for _, details := range errorStatus.Details() {
				switch info := details.(type) {
				case *epb.BadRequest_FieldViolation:
					jsonErr = models.NewBadRequestError(info.Description, info.Field)
				default:
					jsonErr = models.NewInternalServerError(err.Error())
				}
			}
		default:
			jsonErr = models.NewInternalServerError(err.Error())
		}
		return jsonErr
	}

	return nil
}
