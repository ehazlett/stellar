package radiant

type Config struct {
	GRPCAddr     string
	DatastoreUri string
	HTTPPort     int
	HTTPSPort    int
	TLSEmail     string
	Debug        bool
}
