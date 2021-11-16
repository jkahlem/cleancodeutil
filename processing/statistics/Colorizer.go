package statistics

import (
	"returntypes-langserver/processing/typeclasses"
)

type Colorizer struct {
	colorMap map[string]string
}

// Returns the color for a type class as defined in the type class configuration file
func (c *Colorizer) GetColorForTypeClass(typeClassName string) string {
	if c.colorMap == nil {
		c.createColorMap()
	}
	if color, ok := c.colorMap[typeClassName]; ok {
		return color
	}
	return ""
}

// Creates a map for a type class to it's color code
func (c *Colorizer) createColorMap() {
	typeClassConfig := typeclasses.GetTypeClasses()
	c.createColorMapFromTypeClasses(typeClassConfig)
}

func (c *Colorizer) createColorMapFromTypeClasses(config typeclasses.TypeClassConfiguration) {
	c.colorMap = make(map[string]string)
	for _, typeClass := range config.Classes {
		c.colorMap[typeClass.Label] = typeClass.Color
	}
}
