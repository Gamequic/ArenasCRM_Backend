package logsstruct

// Nodo para representar la estructura de logs
type TreeLogNode struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	FileType string        `json:"fileType"`
	Children []TreeLogNode `json:"children,omitempty"`
}
