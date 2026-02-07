package notifications

import "context"

// NotificationStore handles persistence of notifications.
type NotificationStore interface {
	Save(ctx context.Context, notification Notification) error
	SaveBatch(ctx context.Context, notifications []Notification) error
	List(ctx context.Context, userAddress string, opts ListOptions) ([]Notification, int, error)
	MarkAsRead(ctx context.Context, userAddress string, notificationIDs []string) error
}

// DevicePlatform represents push device types.
type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformWeb     DevicePlatform = "web"
)

// DeviceRegistration contains push token metadata.
type DeviceRegistration struct {
	ID         string
	Token      string
	Platform   DevicePlatform
	DeviceName string
	AppVersion string
	CreatedAt  int64
	LastSeenAt *int64
	DisabledAt *int64
}

// DeviceStore handles push device registrations.
type DeviceStore interface {
	RegisterDevice(ctx context.Context, userAddress string, registration DeviceRegistration) (DeviceRegistration, error)
	ListDevices(ctx context.Context, userAddress string) ([]DeviceRegistration, error)
	UpdateDevice(ctx context.Context, userAddress string, registration DeviceRegistration) error
	RemoveDevice(ctx context.Context, userAddress string, deviceID string) error
}
