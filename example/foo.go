package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type UnixMillis int64
type JsonMap map[string]interface{}
type DoubleAlias JsonMap

type InterfaceResponse interface{}
type Instruction bson.M
type BsonID bson.ObjectId

type FooResponse struct {
	BsonID        BsonID                 `json:"bsonId"`
	ID            string                 `json:"id"`
	StartDate     time.Time              `json:"startDate"`
	UniqueId      uuid.UUID              `json:"uniqueId"`
	EndDate       UnixMillis             `json:"endDate"`
	Count         int64                  `json:"count" example:"6"`
	Msg           json.RawMessage        `json:"msg"`
	InnerFoos     []InnerFoo             `json:"foo"`
	Environments  map[string]Environment `json:"environments"`
	FreeForm      interface{}            `json:"freeForm"`
	JsonMap       JsonMap                `json:"jsonMap"`
	DoubleAlias   DoubleAlias            `json:"doubleAlias"`
	InterfaceBlah InterfaceResponse      `json:"interfaceBlah"`
	Instruction   Instruction            `json:"instruction"`
	BsonPtr       *BsonID                `json:"bsonPtr,omitempty" example:"blah blah blah"`
	RandomBool    bool                   `json:"randomBool,omitempty" example:"true"`
	MyEnum        string                 `json:"myEnum" goas:"enum=value1 value2"`
	ChangeReturn  string                 `json:"changeReturn" overrideApiSchemaType:"Instruction"`
}

type FooBody struct {
	Name      string      `json:"name"`
	Example   DoubleAlias `'json:"doubleAlias"`
	Required  bool        `json:"required"`
	ArrayEnum []string    `json:"arrayEnum" goas:"enum=example1 example2"`
}
type Environment struct {
	Name string `json:"name"`
}
type FooPatchOperation struct {
	op    string
	path  string
	value string
}

type FooPatchOperationSet struct {
	operations []FooPatchOperation
}

type FooMergePatch struct {
	Count int64 `json:"count"`
}

// @ApiSchemaName Birb
type Bird struct {
	Animal
	AirspeedVelocity float32 `json:"airspeedVelocity"`
}

// TODO since Animal is not used directly and probably shouldn't be included in the output spec
type Animal struct {
	name string `required:"true"`
}

// @Title Get all foos
// @Tag Foo
// @Description Get all foos
// @OperationId getAllFoos
// @Route /api/v2/foo [get]
// @Success 200 object FooResponse "Successful foo response"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
func getAllFoos() {

}

// @Title Put foo
// @Description Overwrite a foo
// @Route /api/v2/foo [put]
// @Success 200 object FooResponse "Successful foo response"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
func putFoo() {

}

// @Title Post foo
// @Description Post a foo
// @Param foo body FooBody true "New Foo data"
// @Route /api/v2/foo [post]
// @Success 200 object FooResponse "Successful foo response"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
func postFoo() {

}

// @Title Update foo
// @OperationId patchFoo
// @Route /api/v2/foo/{id} [patch]
// @Param id path string true "Foo id"
// @Param foo body oneOf(FooPatchOperationSet,FooMergePatch) true "Foo patch body"
// @Success 204 "No content"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
func patchFoo() {

}

// @Title Get Foo as var
// @Tag Vfoo
// @Description get a foo var
// @Route /api/v2/vfoo [get]
// @Success 200 object FooResponse "Successful foo response"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
var getVarFoo = func() {

}

// @Title Get foos with spaces
// @Tag Foo With Spaces
// @Description Get foos with spaces
// @OperationId getFoosWithSpaces
// @Route /api/v2/foo-spaces [get]
// @Success 200 object FooResponse "Successful foo response"
// @Failure 401 "Invalid access token"
// @Failure 403 "Forbidden"
// @Failure 404 "Invalid resource identifier"
func getFoosSpacesInTag() {

}

// @Title Get all birds
// @OperationId getBirds
// @Route /api/v2/birds [get]
// @Success 200 object Bird "Success"
func getBirds() {

}

// @Title Hidden Route
// @OperationId getHiddenRoutes
// @Route /api/v2/showHidden [get]
// @Success 200 object Bird "Success"
// @Hidden
func getHidden() {

}
