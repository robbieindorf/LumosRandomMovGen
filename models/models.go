package models

type MediaContainer struct {
	Type         string 	   `json:"elementType"`
	Partition    string        `json:"partition"`
	Address      int           `json:"address"`
	MediaType    string        `json:"mediaType"`
	MediaBarcode *string       `json:"mediaBarcode,omitempty"`
}

type Media struct {
	Barcode  string          `json:"barcode"`
	Location *MediaContainer `json:"location,omitempty"`
}

type Magazine struct {
	Barcode string           `json:"barcode"`
	Slots   []MediaContainer `json:"slots"`
}

type MoveRequest struct {
	Source    int 			`json:"source"`
	Dest      int			`json:"dest"`
	Partition string       	`json:"partition,omitempty"`
}

type Move struct {
	ID        int64       `json:"id"`
	Status    string      `json:"status"`
	Type      string      `json:"type"`
	Partition string      `json:"partition"`
	Result    *MoveResult `json:"result,omitempty"`

	// Media Moves
	Media       *Media          `json:"media,omitempty"`
	Destination *MediaContainer `json:"dest,omitempty"`

	// IE Moves
	Taps []string `json:"taps,omitempty"`

	// Export Moves
	Magazines []Magazine `json:"magazines,omitempty"`
}

type MoveResult struct {
	ErrorDesc *string `json:"errorDesc,omitempty"`
	Sense     int     `json:"sense"`
	ASC       int     `json:"asc"`
	ASCQ      int     `json:"ascq"`
}
