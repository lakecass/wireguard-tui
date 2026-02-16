package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name           string
	HeaderBg       lipgloss.Color
	HeaderFg       lipgloss.Color
	ColumnHeaderBg lipgloss.Color // implicit or same as Header? Using separate might be nice
	ColumnHeaderFg lipgloss.Color
	SelectedBg     lipgloss.Color
	SelectedFg     lipgloss.Color
	NormalFg       lipgloss.Color
	DimFg          lipgloss.Color
	KeyBg          lipgloss.Color
	KeyFg          lipgloss.Color
	DescBg         lipgloss.Color
	DescFg         lipgloss.Color
}

var Themes = []Theme{
	// Classic Htop (High Contrast)
	{
		Name:           "Htop Classic",
		HeaderBg:       lipgloss.Color("2"),   // Green
		HeaderFg:       lipgloss.Color("0"),   // Black
		ColumnHeaderBg: lipgloss.Color("0"),   // Black
		ColumnHeaderFg: lipgloss.Color("6"),   // Cyan
		SelectedBg:     lipgloss.Color("6"),   // Cyan
		SelectedFg:     lipgloss.Color("0"),   // Black
		NormalFg:       lipgloss.Color("15"),  // White
		DimFg:          lipgloss.Color("240"), // Gray
		KeyBg:          lipgloss.Color("1"),   // Red
		KeyFg:          lipgloss.Color("0"),   // Black
		DescBg:         lipgloss.Color("2"),   // Green
		DescFg:         lipgloss.Color("0"),   // Black
	},
	// Dracula (Dark Pulse)
	{
		Name:           "Dracula",
		HeaderBg:       lipgloss.Color("62"),  // Purple
		HeaderFg:       lipgloss.Color("255"), // White
		ColumnHeaderBg: lipgloss.Color("236"), // Dark Gray
		ColumnHeaderFg: lipgloss.Color("86"),  // Cyan
		SelectedBg:     lipgloss.Color("44"),  // Cyan/Light Blue
		SelectedFg:     lipgloss.Color("235"), // Dark
		NormalFg:       lipgloss.Color("252"), // Subtex
		DimFg:          lipgloss.Color("60"),  // Comment
		KeyBg:          lipgloss.Color("215"), // Orange
		KeyFg:          lipgloss.Color("235"), // Dark
		DescBg:         lipgloss.Color("62"),  // Purple
		DescFg:         lipgloss.Color("255"), // White
	},
	// Solarized Light
	{
		Name:           "Solarized Light",
		HeaderBg:       lipgloss.Color("136"), // Yellow
		HeaderFg:       lipgloss.Color("230"), // Base3
		ColumnHeaderBg: lipgloss.Color("254"), // Base2
		ColumnHeaderFg: lipgloss.Color("64"),  // Green
		SelectedBg:     lipgloss.Color("33"),  // Blue
		SelectedFg:     lipgloss.Color("255"), // White
		NormalFg:       lipgloss.Color("240"), // Base01
		DimFg:          lipgloss.Color("245"), // Base1
		KeyBg:          lipgloss.Color("166"), // Orange
		KeyFg:          lipgloss.Color("255"), // White
		DescBg:         lipgloss.Color("136"), // Yellow
		DescFg:         lipgloss.Color("230"), // Base3
	},
	// Nord (Arctic)
	{
		Name:           "Nord",
		HeaderBg:       lipgloss.Color("81"),  // Frost Blue
		HeaderFg:       lipgloss.Color("232"), // Dark Black
		ColumnHeaderBg: lipgloss.Color("237"), // Polar Night
		ColumnHeaderFg: lipgloss.Color("81"),  // Frost Blue
		SelectedBg:     lipgloss.Color("88"),  // Frost Red/Aurora (Contrast) -> Actually lets use Frost Blue 81 or 6
		SelectedFg:     lipgloss.Color("232"),
		NormalFg:       lipgloss.Color("255"), // Snow Storm
		DimFg:          lipgloss.Color("243"),
		KeyBg:          lipgloss.Color("81"),
		KeyFg:          lipgloss.Color("232"),
		DescBg:         lipgloss.Color("237"),
		DescFg:         lipgloss.Color("255"),
	},
	// Tokyo Night
	{
		Name:           "Tokyo Night",
		HeaderBg:       lipgloss.Color("111"), // Blue
		HeaderFg:       lipgloss.Color("232"),
		ColumnHeaderBg: lipgloss.Color("236"),
		ColumnHeaderFg: lipgloss.Color("176"), // Pink/Purple
		SelectedBg:     lipgloss.Color("176"), // Magenta
		SelectedFg:     lipgloss.Color("232"),
		NormalFg:       lipgloss.Color("253"),
		DimFg:          lipgloss.Color("240"),
		KeyBg:          lipgloss.Color("111"),
		KeyFg:          lipgloss.Color("232"),
		DescBg:         lipgloss.Color("236"),
		DescFg:         lipgloss.Color("176"),
	},
}
