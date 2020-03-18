package components

type MoveManager interface {
	GenerateMove()
	SendMove()
	CheckMoveStatus()
}
