package middleware

type PoolEntity interface {
	Id() uint32
}

type Pool interface {
	Take() (PoolEntity ,error)
	
}
