package can

type SessionMarshaler interface {
	Marshal(session Session) ([]byte, error)
	Unmarshal([]byte) (Session, error)
}
