package iconst

const (
	Success        = "Y"
	Fail           = "N"
	Timeout        = "Timeout"
	PngFolder      = "xlipboard_png"
	MountFolder    = "xlipboard_fuse"
	MaxMsgSize     = 1024 * 1024 * 32
	MaxLenChanFile = 64
	MaxLenChanConn = 512

	RoutePathSayHello     = "/Sync/SayHello"
	RoutePathSayHowAreYou = "/Sync/SayHowAreYou"
	RoutePathReadDir      = "/Sync/ReadDir"
	RoutePathOpen         = "/Sync/Open"
	RoutePathRelease      = "/Sync/Release"
	RoutePathReadFile     = "/Sync/ReadFile"
	RoutePathStat         = "/Sync/Stat"
)
