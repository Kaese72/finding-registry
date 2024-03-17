package database

import (
	"context"

	"github.com/Kaese72/findings-registry/internal/intermediaries"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBConfig struct {
	ConnectionString string
	DbName           string
}

type mongoFindingsPersistence struct {
	mongoClient *mongo.Client
	dbName      string
}

type ReportLocator struct {
	Type  string `bson:"type"`
	Value string `bson:"value"`
}

func (locator ReportLocator) toIntermediary() intermediaries.ReportLocator {
	return intermediaries.ReportLocator{
		Type:  intermediaries.ReportLocatorType(locator.Type),
		Value: locator.Value,
	}
}

func ReportLocatorFromIntermediary(intermediary intermediaries.ReportLocator) ReportLocator {
	return ReportLocator{
		Type:  string(intermediary.Type),
		Value: intermediary.Value,
	}
}

type ReportDistinguisher struct {
	Type  string `bson:"type"`
	Value string `bson:"value"`
}

func (distinguisher ReportDistinguisher) toIntermediary() intermediaries.ReportDistinguisher {
	return intermediaries.ReportDistinguisher{
		Type:  distinguisher.Type,
		Value: distinguisher.Value,
	}
}

func ReportDistinguisherFromIntermediary(intermediary intermediaries.ReportDistinguisher) ReportDistinguisher {
	return ReportDistinguisher{
		Type:  intermediary.Type,
		Value: intermediary.Value,
	}
}

type Finding struct {
	Identifier            string              `bson:"_id,omitempty"`
	Name                  string              `bson:"name"`
	ReportDistinguisher   ReportDistinguisher `bson:"reportDistinguisher"`
	ReportLocator         ReportLocator       `bson:"reportLocator"`
	ImpliedReportLocators []ReportLocator     `bson:"impliedReportLocators"`
}

func (finding Finding) toIntermediary() intermediaries.Finding {
	implied := []intermediaries.ReportLocator{}
	for index := range finding.ImpliedReportLocators {
		implied = append(implied, finding.ImpliedReportLocators[index].toIntermediary())
	}
	return intermediaries.Finding{
		Identifier:            finding.Identifier,
		Name:                  finding.Name,
		ReportDistinguisher:   finding.ReportDistinguisher.toIntermediary(),
		ReportLocator:         finding.ReportLocator.toIntermediary(),
		ImpliedReportLocators: implied,
	}
}

func findingFromIntermediary(intermediary intermediaries.Finding) Finding {
	reportLocators := []ReportLocator{}
	for index := range intermediary.ImpliedReportLocators {
		reportLocators = append(reportLocators, ReportLocatorFromIntermediary(intermediary.ImpliedReportLocators[index]))
	}
	return Finding{
		Identifier:            intermediary.Identifier,
		Name:                  intermediary.Name,
		ReportDistinguisher:   ReportDistinguisherFromIntermediary(intermediary.ReportDistinguisher),
		ReportLocator:         ReportLocatorFromIntermediary(intermediary.ReportLocator),
		ImpliedReportLocators: reportLocators,
	}
}

func NewMongoFindingsPersistence(config MongoDBConfig) (mongoFindingsPersistence, error) {
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.ConnectionString))
	if err != nil {
		return mongoFindingsPersistence{}, err
	}

	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		return mongoFindingsPersistence{}, err
	}

	return mongoFindingsPersistence{
		mongoClient: mongoClient,
		dbName:      config.DbName,
	}, nil
}

func (persistence mongoFindingsPersistence) findingCollection() *mongo.Collection {
	return persistence.mongoClient.Database(persistence.dbName).Collection("findings")
}

func (persistence mongoFindingsPersistence) UpdateFinding(findingI intermediaries.Finding) (intermediaries.Finding, error) {
	findingC := persistence.findingCollection()
	findingR := Finding{}
	mongoFinding := findingFromIntermediary(findingI)
	err := findingC.FindOneAndUpdate(context.TODO(), bson.D{bson.E{Key: "reportDistinguisher", Value: findingI.ReportDistinguisher}, primitive.E{Key: "reportLocator", Value: mongoFinding.ReportLocator}}, bson.M{"$set": mongoFinding}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&findingR)
	return findingR.toIntermediary(), err
}

func (persistence mongoFindingsPersistence) GetFinding(identifier string) (intermediaries.Finding, error) {
	findinfC := persistence.findingCollection()
	objID, _ := primitive.ObjectIDFromHex(identifier)
	findingR := Finding{}
	err := findinfC.FindOne(context.TODO(), bson.D{primitive.E{Key: "_id", Value: objID}}).Decode(&findingR)
	return findingR.toIntermediary(), err
}

func (persistence mongoFindingsPersistence) GetFindings() ([]intermediaries.Finding, error) {
	findinfC := persistence.findingCollection()
	findingIs := []intermediaries.Finding{}
	cursor, err := findinfC.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.TODO()) {
		findingR := Finding{}
		err := cursor.Decode(&findingR)
		if err != nil {
			panic(err.Error())
		}
		findingIs = append(findingIs, findingR.toIntermediary())
	}
	return findingIs, err
}

func (persistence mongoFindingsPersistence) Purge() error {
	findingC := persistence.findingCollection()
	_, err := findingC.DeleteMany(context.TODO(), bson.D{})
	return err
}
