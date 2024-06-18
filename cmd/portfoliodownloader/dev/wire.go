// go:build wireinject
//go:build wireinject
// +build wireinject

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

package main

import (
	"gorm.io/gorm"

	"github.com/dhojayev/traderepublic-portfolio-downloader/cmd/portfoliodownloader"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api/auth"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api/timeline/activitylog"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api/timeline/details"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api/timeline/transactions"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/api/websocket"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/console"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/database"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/filesystem"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/portfolio/activity"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/portfolio/document"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/portfolio/transaction"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/reader"
	"github.com/dhojayev/traderepublic-portfolio-downloader/internal/writer"

	"github.com/google/wire"
	log "github.com/sirupsen/logrus"
)

var (
	DefaultSet = wire.NewSet(
		portfoliodownloader.NewApp,
		transactions.NewClient,
		transactions.NewEventTypeResolver,
		details.NewClient,
		details.NewTypeResolver,
		transaction.NewModelBuilderFactory,
		document.NewModelBuilder,
		database.NewSQLiteOnFS,
		transaction.NewCSVEntryFactory,
		filesystem.NewCSVReader,
		filesystem.NewCSVWriter,
		transaction.NewProcessor,
		document.NewDownloader,
		document.NewDateResolver,
		ProvideTransactionRepository,
		ProvideInstrumentRepository,
		ProvideDocumentRepository,
		activitylog.NewClient,
		activity.NewProcessor,
		activity.NewHandler,
		transaction.NewHandler,

		wire.Bind(new(transactions.ClientInterface), new(transactions.Client)),
		wire.Bind(new(transactions.EventTypeResolverInterface), new(transactions.EventTypeResolver)),
		wire.Bind(new(details.ClientInterface), new(details.Client)),
		wire.Bind(new(details.TypeResolverInterface), new(details.TypeResolver)),
		wire.Bind(new(transaction.ProcessorInterface), new(transaction.Processor)),
		wire.Bind(new(transaction.ModelBuilderFactoryInterface), new(transaction.ModelBuilderFactory)),
		wire.Bind(new(document.ModelBuilderInterface), new(document.ModelBuilder)),
		wire.Bind(new(transaction.RepositoryInterface), new(*database.Repository[*transaction.Model])),
		wire.Bind(new(transaction.InstrumentRepositoryInterface), new(*database.Repository[*transaction.Instrument])),
		wire.Bind(new(document.DownloaderInterface), new(document.Downloader)),
		wire.Bind(new(document.DateResolverInterface), new(document.DateResolver)),
		wire.Bind(new(document.RepositoryInterface), new(*database.Repository[*document.Model])),
		wire.Bind(new(filesystem.CSVReaderInterface), new(filesystem.CSVReader)),
		wire.Bind(new(filesystem.CSVWriterInterface), new(filesystem.CSVWriter)),
		wire.Bind(new(activitylog.ClientInterface), new(activitylog.Client)),
		wire.Bind(new(activity.ProcessorInterface), new(activity.Processor)),
		wire.Bind(new(activity.HandlerInterface), new(activity.Handler)),
		wire.Bind(new(transaction.HandlerInterface), new(transaction.Handler)),
	)

	RemoteSet = wire.NewSet(
		DefaultSet,
		api.NewClient,
		auth.NewClient,
		console.NewAuthService,
		websocket.NewReader,
		filesystem.NewJSONWriter,

		wire.Bind(new(auth.ClientInterface), new(*auth.Client)),
		wire.Bind(new(console.AuthServiceInterface), new(*console.AuthService)),
		wire.Bind(new(reader.Interface), new(*websocket.Reader)),
		wire.Bind(new(writer.Interface), new(*filesystem.JSONWriter)),
	)

	LocalSet = wire.NewSet(
		DefaultSet,
		writer.NewNilWriter,
		reader.NewJSONReader,

		wire.Bind(new(reader.Interface), new(*reader.JSONReader)),
		wire.Bind(new(writer.Interface), new(writer.NilWriter)),
	)
)

func CreateLocalApp(baseDir string, logger *log.Logger) (portfoliodownloader.App, error) {
	wire.Build(LocalSet)

	return portfoliodownloader.App{}, nil
}

func CreateRemoteApp(logger *log.Logger) (portfoliodownloader.App, error) {
	wire.Build(RemoteSet)

	return portfoliodownloader.App{}, nil
}

func ProvideTransactionRepository(db *gorm.DB, logger *log.Logger) (*database.Repository[*transaction.Model], error) {
	return database.NewRepository[*transaction.Model](db, logger)
}

func ProvideInstrumentRepository(db *gorm.DB, logger *log.Logger) (*database.Repository[*transaction.Instrument], error) {
	return database.NewRepository[*transaction.Instrument](db, logger)
}

func ProvideDocumentRepository(db *gorm.DB, logger *log.Logger) (*database.Repository[*document.Model], error) {
	return database.NewRepository[*document.Model](db, logger)
}
