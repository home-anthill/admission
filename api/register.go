package api

import (
	"admission/api/grpc/register"
	"admission/customerrors"
	"admission/db"
	"admission/grpcutil"
	"admission/httputil"
	"admission/models"
	"admission/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// SpecListItemReq represents a feature spec list item in a device registration request.
type SpecListItemReq struct {
	Value *int   `json:"value" validate:"required"`
	Text  string `json:"text" validate:"required,alphanum,min=1"`
}

// SpecReq represents a feature spec in a device registration request.
type SpecReq struct {
	Format models.SpecFormat `json:"format" validate:"required,oneof=bool int float list"`
	Min    *float64          `json:"min,omitempty"`
	Max    *float64          `json:"max,omitempty"`
	Step   *float64          `json:"step,omitempty"`
	List   []SpecListItemReq `json:"list,omitempty" validate:"max=20,dive"`
}

// FeatureReq represents a single feature in a device registration request.
type FeatureReq struct {
	Type   models.Type `json:"type" validate:"required,oneof='controller' 'sensor'"`
	Name   string      `json:"name" validate:"required,min=2,max=20,alphanum"`
	Enable bool        `json:"enable" validate:"boolean"`
	Order  int         `json:"order" validate:"required,gte=1"`
	Unit   string      `json:"unit" validate:"required,min=1,max=10"`
	Spec   SpecReq     `json:"spec" validate:"required"`
}

// DeviceRegisterReq is the expected JSON body for a device registration request.
type DeviceRegisterReq struct {
	Mac          string       `json:"mac" validate:"required,mac"`
	Manufacturer string       `json:"manufacturer" validate:"required,min=3,max=50"`
	Model        string       `json:"model" validate:"required,min=3,max=20"`
	APIToken     string       `json:"apiToken" validate:"required,uuid4"`
	Features     []FeatureReq `json:"features" validate:"required,max=16,dive"`
}

// DeviceRegisterRes is the response returned after a successful device registration.
type DeviceRegisterRes struct {
	UUID         string           `json:"uuid"`
	Mac          string           `json:"mac"`
	Manufacturer string           `json:"manufacturer"`
	Model        string           `json:"model"`
	Features     []models.Feature `json:"features"`
}

// SensorRegisterReq is the payload sent to the downstream sensor registration service.
type SensorRegisterReq struct {
	DeviceUuid     string `json:"deviceUuid"`
	Mac            string `json:"mac"`
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	ProfileOwnerID string `json:"profileOwnerId"`
	APIToken       string `json:"apiToken"`
	FeatureUUID    string `json:"featureUuid"`
}

// Register handles device registration via REST, gRPC, and MongoDB.
type Register struct {
	client             *mongo.Client
	collDevices        *mongo.Collection
	collProfiles       *mongo.Collection
	logger             *zap.SugaredLogger
	grpcTarget         string
	keepAliveSensorURL string
	registerSensorURL  string
	validate           *validator.Validate
}

// NewRegister creates a new Register handler with the given dependencies.
func NewRegister(logger *zap.SugaredLogger, client *mongo.Client, validate *validator.Validate) *Register {
	grpcURL := os.Getenv("GRPC_URL")
	sensorServerURL := os.Getenv("HTTP_SENSOR_SERVER") + ":" + os.Getenv("HTTP_SENSOR_PORT")
	keepAliveSensorURL := sensorServerURL + os.Getenv("HTTP_SENSOR_KEEPALIVE_API")
	registerSensorURL := sensorServerURL + os.Getenv("HTTP_SENSOR_REGISTER_API")

	return &Register{
		client:             client,
		collDevices:        db.GetCollections(client).Devices,
		collProfiles:       db.GetCollections(client).Profiles,
		logger:             logger,
		grpcTarget:         grpcURL,
		keepAliveSensorURL: keepAliveSensorURL,
		registerSensorURL:  registerSensorURL,
		validate:           validate,
	}
}

// PostRegister handles device registration requests.
func (handler *Register) PostRegister(c *gin.Context) {
	handler.logger.Info("REST - PostRegister called")
	ctx := c.Request.Context()

	var registerBody DeviceRegisterReq
	if err := c.ShouldBindJSON(&registerBody); err != nil {
		handler.logger.Errorw("REST - PostRegister - Cannot bind request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := handler.validate.Struct(registerBody)
	if err != nil {
		handler.logger.Errorw("REST - PostRegister - request body is not valid", "error", err)
		errFields := utils.GetErrorMessage(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, these fields are not valid:" + errFields})
		return
	}

	apiTokenHash, err := utils.HashAPIToken(registerBody.APIToken)
	if err != nil {
		handler.logger.Errorw("REST - PostRegister - Cannot hash profile apiToken", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register device"})
		return
	}

	// search if profile token exists and retrieve profile
	var profileFound models.Profile
	errProfile := handler.collProfiles.FindOne(ctx, bson.M{
		"apiTokenHash": apiTokenHash,
	}).Decode(&profileFound)
	if errProfile != nil {
		if !errors.Is(errProfile, mongo.ErrNoDocuments) {
			handler.logger.Errorw("REST - PostRegister - Cannot query profile with that apiToken", "error", errProfile)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register device"})
			return
		}
		handler.logger.Errorw("REST - PostRegister - Cannot find profile with that apiToken", "error", errProfile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot register, profile token missing or not valid"})
		return
	}

	// search and skip db add if device already exists
	var device models.Device
	err = handler.collDevices.FindOne(ctx, bson.M{
		"mac": registerBody.Mac,
	}).Decode(&device)
	if err == nil {
		if profileOwnsDevice(&profileFound, device.ID) {
			handler.logger.Info("REST - PostRegister - Device already registered for this profile")
		} else {
			handler.logger.Warn("REST - PostRegister - Device already registered for another profile")
		}
		c.JSON(http.StatusConflict, gin.H{"message": "Already registered"})
		return
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		handler.logger.Errorw("REST - PostRegister - Cannot query existing device by mac", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register device"})
		return
	}

	insertDate := time.Now()
	device = models.Device{
		ID:   bson.NewObjectID(),
		UUID: uuid.NewString(),
		Mac:  registerBody.Mac,
		// init by default Name with MAC address
		Name:         registerBody.Mac,
		Manufacturer: registerBody.Manufacturer,
		Model:        registerBody.Model,
		CreatedAt:    insertDate,
		ModifiedAt:   insertDate,
		Features: utils.MapSlice(registerBody.Features, func(fReq FeatureReq) models.Feature {
			return models.Feature{
				UUID:   uuid.NewString(),
				Type:   fReq.Type,
				Name:   fReq.Name,
				Enable: fReq.Enable,
				Order:  fReq.Order,
				Unit:   fReq.Unit,
				Spec:   specReqToModel(fReq.Spec),
			}
		}),
	}

	controllers := utils.Filter(device.Features, func(f models.Feature) bool { return f.Type == models.Controller })
	sensors := utils.Filter(device.Features, func(f models.Feature) bool { return f.Type == models.Sensor })
	handler.logger.Debugf("REST - PostRegister - controllers %v", controllers)
	handler.logger.Debugf("REST - PostRegister - sensors %v", sensors)

	// register controllers via gRPC
	if len(controllers) > 0 {
		errRegister := handler.registerControllersViaGRPC(ctx, &device, controllers, &profileFound, registerBody.APIToken)
		if errRegister != nil {
			handler.logger.Errorw("REST - PostRegister - cannot register controller device via gRPC", "error", errRegister)
			if re, ok := errRegister.(*customerrors.ErrorWrapper); ok {
				handler.logger.Errorw("REST - PostRegister - cannot register device", "status", re.Code, "message", re.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register controller device"})
			return
		}
		handler.logger.Debug("REST - PostRegister - controller devices registered")
	}

	// register sensors via REST
	if len(sensors) > 0 {
		errRegister := handler.registerSensorsViaHTTP(&device, sensors, &profileFound, registerBody.APIToken)
		if errRegister != nil {
			handler.logger.Errorw("REST - PostRegister - cannot register sensor device via HTTP", "error", errRegister)
			if re, ok := errRegister.(*customerrors.ErrorWrapper); ok {
				handler.logger.Errorw("REST - PostRegister - cannot register device", "status", re.Code, "message", re.Message)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register sensor device"})
			return
		}
		handler.logger.Debug("REST - PostRegister - sensor devices registered successfully")
	}

	// Insert device into admission database
	errInsDb := handler.insertDevice(ctx, &device, &profileFound)
	if errInsDb != nil {
		var wrapped customerrors.ErrorWrapper
		if errors.As(errInsDb, &wrapped) && wrapped.Code == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{"message": "Already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot register device"})
		return
	}

	handler.logger.Debugf("REST - PostRegister - registered device = %#v", device)
	c.JSON(http.StatusOK, DeviceRegisterRes{
		UUID:         device.UUID,
		Mac:          device.Mac,
		Manufacturer: device.Manufacturer,
		Model:        device.Model,
		Features:     device.Features,
	})
}

func (handler *Register) registerSensorsViaHTTP(device *models.Device, sensorFeatures []models.Feature, profileFound *models.Profile, apiToken string) error {
	// check if service is available calling keep-alive
	// TODO remove this in a production code
	statusCode, _, keepAliveErr := httputil.Get(handler.keepAliveSensorURL)
	if keepAliveErr != nil {
		return customerrors.Wrap(http.StatusInternalServerError, keepAliveErr, "Cannot call keepAlive of remote register service")
	}
	if statusCode < 200 || statusCode >= 300 {
		return customerrors.Wrap(statusCode, nil, "keepAlive of remote register service returned non-2xx status")
	}

	for _, feature := range sensorFeatures {
		payload := SensorRegisterReq{
			DeviceUuid:     device.UUID,
			Mac:            device.Mac,
			Manufacturer:   device.Manufacturer,
			Model:          device.Model,
			ProfileOwnerID: profileFound.ID.Hex(),
			APIToken:       apiToken,
			FeatureUUID:    feature.UUID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create payload to register sensor service")
		}
		// do the real call to the remote registration service
		sc, _, err := httputil.Post(handler.registerSensorURL+feature.Name, payloadJSON)
		if err != nil {
			return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot register sensor device feature "+feature.Name)
		}
		if sc < 200 || sc >= 300 {
			return customerrors.Wrap(sc, nil, "Remote sensor service returned non-2xx status for feature "+feature.Name)
		}
	}
	return nil
}

func (handler *Register) registerControllersViaGRPC(ctx context.Context, device *models.Device, controllerFeatures []models.Feature, profileFound *models.Profile, apiToken string) error {
	handler.logger.Info("gRPC - registerControllersViaGRPC - Sending register via gRPC...")
	// Set up a connection to the gRPC server.
	securityDialOption, isSecure, err := grpcutil.BuildSecurityDialOption()
	if err != nil {
		return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create securityDialOption to prepare the gRPC connection")
	}
	if isSecure {
		handler.logger.Debug("registerControllersViaGRPC - GRPC secure enabled!")
	} else {
		handler.logger.Info("registerControllersViaGRPC - GRPC secure NOT enabled!")
	}

	conn, err := grpc.NewClient(handler.grpcTarget, securityDialOption)
	if err != nil {
		handler.logger.Error("gRPC - registerControllersViaGRPC - cannot connect via gRPC", err)
		return customerrors.GrpcSendError{
			Status:  customerrors.ConnectionError,
			Message: "Cannot connect to api-devices",
		}
	}
	defer conn.Close()
	client := register.NewRegistrationClient(conn)

	// Use a per-call timeout so each RPC gets its own deadline
	for _, feature := range controllerFeatures {
		callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := client.Register(callCtx, &register.RegisterRequest{
			DeviceUuid:     device.UUID,
			Mac:            device.Mac,
			Manufacturer:   device.Manufacturer,
			Model:          device.Model,
			ProfileOwnerId: profileFound.ID.Hex(),
			ApiToken:       apiToken,
			Feature: &register.RegisterFeature{
				FeatureUuid: feature.UUID,
				FeatureName: feature.Name,
			},
		})
		cancel()
		if err != nil {
			handler.logger.Errorw("gRPC - registerControllersViaGRPC - cannot invoke Register via gRPC", "error", err)
			return customerrors.Wrap(http.StatusInternalServerError, err, "Cannot invoke Register via gRPC")
		}
	}

	return nil
}

func (handler *Register) insertDevice(ctx context.Context, device *models.Device, profile *models.Profile) error {
	// start-session
	dbSession, err := handler.client.StartSession()
	if err != nil {
		handler.logger.Errorw("insertDevice - cannot start a db session", "error", err)
		return customerrors.Wrap(http.StatusInternalServerError, err, "unknown error while trying to register a device")
	}
	// Defers ending the session after the transaction is committed or ended
	defer dbSession.EndSession(ctx)

	_, errTrans := dbSession.WithTransaction(ctx, func(sessionCtx context.Context) (any, error) {
		// Official `mongo-driver` documentation state: "callback may be run
		// multiple times during WithTransaction due to retry attempts, so it must be idempotent."

		// Insert device
		_, errInsert := handler.collDevices.InsertOne(sessionCtx, device)
		if errInsert != nil {
			if mongo.IsDuplicateKeyError(errInsert) {
				return nil, customerrors.Wrap(http.StatusConflict, errInsert, "Device already registered")
			}
			return nil, customerrors.Wrap(http.StatusInternalServerError, errInsert, "Cannot insert the new device")
		}
		// push device.ID to profile.devices into admission database
		_, errUpd := handler.collProfiles.UpdateOne(
			sessionCtx,
			bson.M{"_id": profile.ID},
			bson.M{"$addToSet": bson.M{"devices": device.ID}},
		)
		if errUpd != nil {
			return nil, customerrors.Wrap(http.StatusInternalServerError, errUpd, "Cannot update profile with the new device")
		}
		return nil, nil
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))

	if errTrans != nil {
		handler.logger.Errorw("insertDevice - insert device in transaction", "error", errTrans)
	}
	return errTrans
}

func specReqToModel(spec SpecReq) models.Spec {
	modelSpec := models.Spec{
		Format: spec.Format,
		Min:    spec.Min,
		Max:    spec.Max,
		Step:   spec.Step,
	}
	if len(spec.List) > 0 {
		modelSpec.List = utils.MapSlice(spec.List, func(item SpecListItemReq) models.SpecListItem {
			return models.SpecListItem{
				Value: float32(*item.Value),
				Text:  item.Text,
			}
		})
	}
	return modelSpec
}

func profileOwnsDevice(profile *models.Profile, deviceID bson.ObjectID) bool {
	for _, profileDeviceID := range profile.Devices {
		if profileDeviceID == deviceID {
			return true
		}
	}
	return false
}
