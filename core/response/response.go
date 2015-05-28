package response

import (
	. "github.com/zubairhamed/go-lwm2m/api"
	"github.com/zubairhamed/go-lwm2m/core/values"
	. "github.com/zubairhamed/goap"
)

// 201 Created
func Created() Response {
	return &CreatedResponse{}
}

type CreatedResponse struct {
}

func (r *CreatedResponse) GetResponseCode() CoapCode {
	return COAPCODE_201_CREATED
}

func (r *CreatedResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 202 Deleted
func Deleted() Response {
	return &DeletedResponse{}
}

type DeletedResponse struct {
}

func (r *DeletedResponse) GetResponseCode() CoapCode {
	return COAPCODE_202_DELETED
}

func (r *DeletedResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 204 Changed
func Changed() Response {
	return &ChangedResponse{}
}

type ChangedResponse struct {
}

func (r *ChangedResponse) GetResponseCode() CoapCode {
	return COAPCODE_204_CHANGED
}

func (r *ChangedResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 205 Content
func Content(val ResponseValue) Response {
	return &ContentResponse{
		val: val,
	}
}

type ContentResponse struct {
	val ResponseValue
}

func (r *ContentResponse) GetResponseCode() CoapCode {
	return COAPCODE_205_CONTENT
}

func (r *ContentResponse) GetResponseValue() ResponseValue {
	return r.val
}

// 400 Bad Request
func BadRequest() Response {
	return &BadRequestResponse{}
}

type BadRequestResponse struct {
}

func (r *BadRequestResponse) GetResponseCode() CoapCode {
	return COAPCODE_400_BAD_REQUEST
}

func (r *BadRequestResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 401 Unauthorized
func Unauthorized() Response {
	return &UnauthorizedResponse{}
}

type UnauthorizedResponse struct {
}

func (r *UnauthorizedResponse) GetResponseCode() CoapCode {
	return COAPCODE_401_UNAUTHORIZED
}

func (r *UnauthorizedResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 404 Not Found
func NotFound() Response {
	return &NotFoundResponse{}
}

type NotFoundResponse struct {
}

func (r *NotFoundResponse) GetResponseCode() CoapCode {
	return COAPCODE_404_NOT_FOUND
}

func (r *NotFoundResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 405 Method Not Allowed
func MethodNotAllowed() Response {
	return &MethodNotAllowedResponse{}
}

type MethodNotAllowedResponse struct {
}

func (r *MethodNotAllowedResponse) GetResponseCode() CoapCode {
	return COAPCODE_405_METHOD_NOT_ALLOWED
}

func (r *MethodNotAllowedResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}

// 409 Conflict
func Conflict() Response {
	return &ConflictResponse{}
}

type ConflictResponse struct {
}

func (r *ConflictResponse) GetResponseCode() CoapCode {
	return COAPCODE_409_CONFLICT
}

func (r *ConflictResponse) GetResponseValue() ResponseValue {
	return values.Empty()
}