package adb

type AdbTool struct {
	controller *Controller
}

func NewTools(serial string) []string {
	// return &Controller{
	// 	Serial:        serial,
	// 	DeviceManager: adb.NewManager(""),
	// 	Memory:        make([]string, 0),
	// 	Screenshots:   make([]ScreenshotInfo, 0),
	// }
	adbTool := &AdbTool{
		controller: NewController(serial),
	}

}
