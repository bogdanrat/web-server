package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bogdanrat/web-server/service/auth/lib"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"image/png"
	"log"
)

type AuthServer struct{}

func (s *AuthServer) GenerateQRCode(ctx context.Context, req *pb.GenerateQRCodeRequest) (*pb.GenerateQRCodeResponse, error) {
	// error handling
	if !lib.IsValidEmail(req.Email) {
		errorStatus := status.New(codes.InvalidArgument, "invalid email format")
		details, err := errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
			Field:       "Email",
			Description: fmt.Sprintf("Email %s is not valid", req.Email),
		})
		if err != nil {
			return nil, errorStatus.Err()
		}
		return nil, details.Err()
	}

	img, secret, err := lib.GenerateQRCode(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating qr code: %s", err)
	}

	buffer := &bytes.Buffer{}
	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	// detect when the client has reached the deadline specified when invoking the RPC
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s\n", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.GenerateQRCodeResponse{
		Image:  buffer.Bytes(),
		Secret: secret,
	}, status.New(codes.OK, "").Err()
}

func (s *AuthServer) ValidateQRCode(ctx context.Context, req *pb.ValidateQRCodeRequest) (*pb.ValidateQRCodeResponse, error) {
	authenticated, err := lib.ValidateQRCode(req.QrCode, req.QrSecret)
	if err != nil {
		errorStatus := status.New(codes.InvalidArgument, err.Error())
		details, err := errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
			Field:       "Email",
			Description: fmt.Sprintf("QR Code %s is not valid", req.QrCode),
		})
		if err != nil {
			return nil, errorStatus.Err()
		}
		return nil, details.Err()
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s\n", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.ValidateQRCodeResponse{Validated: authenticated}, status.New(codes.OK, "").Err()
}

func (s *AuthServer) GenerateToken(ctx context.Context, req *pb.GenerateTokenRequest) (*pb.GenerateTokenResponse, error) {
	token, err := lib.GenerateToken(req.Email, req.AccessTokenDuration, req.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not generate token: %s", err)
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s\n", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.GenerateTokenResponse{Token: token}, status.New(codes.OK, "").Err()
}

func (s *AuthServer) ValidateAccessToken(ctx context.Context, req *pb.ValidateAccessTokenRequest) (*pb.ValidateAccessTokenResponse, error) {
	claims, err := lib.ValidateAccessToken(req.SignedToken)
	if err != nil {
		errorStatus, _ := status.FromError(err)
		var details *status.Status
		switch errorStatus.Code() {
		case codes.InvalidArgument:
			details, err = errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
				Field:       "SignedToken",
				Description: "Invalid JWT format",
			})

		case codes.PermissionDenied:
			details, err = errorStatus.WithDetails(&epb.PreconditionFailure{
				Violations: []*epb.PreconditionFailure_Violation{
					{Description: err.Error()},
				},
			})
		default:
			return nil, err
		}
		if err != nil {
			return nil, errorStatus.Err()
		}
		return nil, details.Err()
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s\n", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.ValidateAccessTokenResponse{
		Email:      claims.Email,
		AccessUuid: claims.AccessUUID,
	}, status.New(codes.OK, "").Err()
}

func (s *AuthServer) ValidateRefreshToken(ctx context.Context, req *pb.ValidateRefreshTokenRequest) (*pb.ValidateRefreshTokenResponse, error) {
	claims, err := lib.ValidateRefreshToken(req.SignedToken)
	if err != nil {
		if errorStatus, _ := status.FromError(err); errorStatus.Code() == codes.InvalidArgument {
			details, err := errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
				Field:       "SignedToken",
				Description: "Invalid JWT format",
			})
			if err != nil {
				return nil, errorStatus.Err()
			}
			return nil, details.Err()
		}
		return nil, err
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state: %s\n", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.ValidateRefreshTokenResponse{
		Email:       claims.Email,
		RefreshUuid: claims.RefreshUUID,
	}, nil
}
