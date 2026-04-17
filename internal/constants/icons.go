package constants

// ValidTypeIcons lists the allowed icon names for collection and asset types.
var ValidTypeIcons = map[string]bool{
	"bezier": true, "bone": true, "book": true, "boxes": true, "bulb": true,
	"camera-flash": true, "camera": true, "clapboard": true, "compass": true, "cube": true,
	"drum": true, "film-reel": true, "film-strip": true, "fire": true, "flow-chart": true,
	"four-squares": true, "home": true, "image": true, "lamp": true, "link": true,
	"man-running": true, "masks": true, "music": true, "mystery-ball": true, "open-book": true,
	"package": true, "palette": true, "scissors": true, "shapes": true, "stall": true,
	"texture": true, "tree": true, "video-camera": true, "website": true,
}

// ValidTypeIconsList returns the valid icon names as a slice for iteration.
func ValidTypeIconsList() []string {
	icons := make([]string, 0, len(ValidTypeIcons))
	for icon := range ValidTypeIcons {
		icons = append(icons, icon)
	}
	return icons
}

// CollectionTypeIcons lists icons commonly used for collection types.
var CollectionTypeIcons = []string{
	"boxes", "film-reel", "clapboard", "video-camera", "bone", "cube", "tree", "four-squares", "package", "shapes", "compass", "home",
}

// AssetTypeIcons lists icons commonly used for asset types.
var AssetTypeIcons = []string{
	"man-running", "music", "palette", "film-strip", "image", "scissors", "bulb",
	"fire", "camera", "stall", "lamp", "texture", "bezier", "camera-flash",
	"mystery-ball", "drum", "flow-chart", "book", "masks", "open-book", "website", "link",
}
