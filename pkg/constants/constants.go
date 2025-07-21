package constants

const (
	GrpcPort                   = "GRPC_PORT"
	HttpPort                   = "HTTP_PORT"
	ConfigPath                 = "CONFIG_PATH"
	KafkaBrokers               = "KAFKA_BROKERS"
	JaegerHostPort             = "JAEGER_HOST"
	RedisAddr                  = "REDIS_ADDR"
	MongoDbURI                 = "MONGO_URI"
	EventStoreConnectionString = "EVENT_STORE_CONNECTION_STRING"
	ElasticUrl                 = "ELASTIC_URL"

	ReaderServicePort = "READER_SERVICE"

	Yaml          = "yaml"
	Tcp           = "tcp"
	Redis         = "redis"
	Kafka         = "kafka"
	Mongo         = "mongo"
	Scylla        = "scylla"
	ElasticSearch = "elasticSearch"

	GRPC     = "GRPC"
	SIZE     = "SIZE"
	URI      = "URI"
	STATUS   = "STATUS"
	HTTP     = "HTTP"
	ERROR    = "ERROR"
	METHOD   = "METHOD"
	METADATA = "METADATA"
	REQUEST  = "REQUEST"
	REPLY    = "REPLY"
	TIME     = "TIME"

	Topic        = "topic"
	Partition    = "partition"
	Message      = "message"
	WorkerID     = "workerID"
	Offset       = "offset"
	Time         = "time"
	GroupName    = "GroupName"
	StreamID     = "StreamID"
	EventID      = "EventID"
	EventType    = "EventType"
	EventNumber  = "EventNumber"
	CreatedDate  = "CreatedDate"
	UserMetadata = "UserMetadata"

	Page   = "page"
	Size   = "size"
	Search = "search"
	ID     = "id"

	EsAll = "$all"

	Validate        = "validate"
	FieldValidation = "field validation"
	RequiredHeaders = "required header"
	Base64          = "base64"
	Unmarshal       = "unmarshal"
	Uuid            = "uuid"
	Cookie          = "cookie"
	Token           = "token"
	Bcrypt          = "bcrypt"
	SQLState        = "sqlstate"

	MongoProjection   = "(MongoDB Projection)"
	ElasticProjection = "(Elastic Projection)"

	ProductId     = "_id"
	ProductName   = "product_name"
	OriginalPrice = "original_price"

	DiscountPromotion = "discount_promotion"
	DiscountPrice     = "discount_price"
	DiscountOff       = "discount_off"
	PurchaseLimit     = "purchase_limit"
	PromotionStock    = "promotion_stock"
	DiscountEnabled   = "discount_enabled"

	PromptVisualImage  = "prompt_visual_image"
	CoverImage         = "cover_image"
	Images             = "images"
	VideoUrl           = "video_url"
	ProductDescription = "product_description"
	ProductURL         = "product_url"
	CategoryID         = "category_id"
	TopicID            = "topic_id"

	Variations    = "variations"
	VariationName = "variation_name"
	Option        = "option"
	Price         = "price"
	Stock         = "stock"
	Image         = "image"

	Specifications = "specifications"
	AttributeName  = "attribute_name"
	Value          = "value"
	Unit           = "unit"

	UsageConfig      = "usage_config"
	NumberOfUses     = "number_of_uses"
	MinimumUsageTime = "minimum_usage_time"
	MaximumUsageTime = "maximum_usage_time"

	UserID = "user_id"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	TokenKey = contextKey("token")
)
