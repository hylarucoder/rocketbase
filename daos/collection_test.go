package daos_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hylarucoder/rocketbase/daos"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/models/schema"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/hylarucoder/rocketbase/tools/list"
	"github.com/hylarucoder/rocketbase/tools/types"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *CollectionTestSuite) TestCollectionQuery() {
	app := suite.App
	expected := "SELECT {{_collections}}.* FROM \"_collections\""

	sql := app.Dao().CollectionQuery().Build().SQL()
	assert.Equal(suite.T(), expected, sql)
}

func (suite *CollectionTestSuite) TestFindCollectionsByType() {
	app := suite.App

	scenarios := []struct {
		collectionType string
		expectError    bool
		expectTotal    int
	}{
		{"", false, 0},
		{"unknown", false, 0},
		{models.CollectionTypeAuth, false, 3},
		{models.CollectionTypeBase, false, 5},
	}

	for i, scenario := range scenarios {
		collections, err := app.Dao().FindCollectionsByType(scenario.collectionType)

		assert.Equal(suite.T(), scenario.expectError, err != nil, "[%d] Expected hasErr to be %v, got %v (%v)", i, scenario.expectError, err != nil, err)

		assert.Equal(suite.T(), scenario.expectTotal, len(collections), "[%d] Expected %d collections, got %d", i, scenario.expectTotal, len(collections))

		for _, c := range collections {
			assert.Equal(suite.T(), scenario.collectionType, c.Type, "[%d] Expected collection with type %s, got %s: \n%v", i, scenario.collectionType, c.Type, c)
		}
	}
}

func (suite *CollectionTestSuite) TestFindCollectionByNameOrId() {
	app := suite.App

	scenarios := []struct {
		nameOrId    string
		expectError bool
	}{
		{"", true},
		{"missing", true},
		{"2108348993330216960", false},
		{"demo1", false},
		{"DEMO1", false}, // case insensitive check
	}

	for i, scenario := range scenarios {
		model, err := app.Dao().FindCollectionByNameOrId(scenario.nameOrId)

		hasErr := err != nil
		assert.Equal(suite.T(), scenario.expectError, hasErr, "[%d] Expected hasErr to be %v, got %v (%v)", i, scenario.expectError, hasErr, err)

		if model != nil && model.Id != scenario.nameOrId && !strings.EqualFold(model.Name, scenario.nameOrId) {
			suite.T().Errorf("[%d] Expected model with identifier %s, got %v", i, scenario.nameOrId, model)
		}
	}
}

func (suite *CollectionTestSuite) TestIsCollectionNameUnique() {
	app := suite.App
	scenarios := []struct {
		name      string
		excludeId string
		expected  bool
	}{
		{"", "", false},
		{"demo1", "", false},
		{"Demo1", "", false},
		{"new", "", true},
		{"demo1", "2108348993330216960", true},
	}

	for i, scenario := range scenarios {
		result := app.Dao().IsCollectionNameUnique(scenario.name, scenario.excludeId)
		assert.Equal(suite.T(), scenario.expected, result, "[%d] Expected %v, got %v", i, scenario.expected, result)
	}
}

func (suite *CollectionTestSuite) TestFindCollectionReferences() {
	app := suite.App

	collection, err := app.Dao().FindCollectionByNameOrId("demo3")
	assert.Nil(suite.T(), err)

	result, err := app.Dao().FindCollectionReferences(
		collection,
		collection.Id,
		// test whether "nonempty" exclude ids condition will be skipped
		"",
		"",
	)
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), 1, len(result))

	expectedFields := []string{
		"rel_one_no_cascade",
		"rel_one_no_cascade_required",
		"rel_one_cascade",
		"rel_many_no_cascade",
		"rel_many_no_cascade_required",
		"rel_many_cascade",
	}

	for col, fields := range result {
		assert.Equal(suite.T(), "demo4", col.Name)

		assert.Equal(suite.T(), len(fields), len(expectedFields))
		for i, f := range fields {
			assert.True(suite.T(), list.ExistInSlice(f.Name, expectedFields), "[%d] Didn't expect field %v", i, f)
		}
	}
}

func (suite *CollectionTestSuite) TestDeleteCollection() {
	app := suite.App

	colUnsaved := &models.Collection{}

	colAuth, err := app.Dao().FindCollectionByNameOrId("users")
	assert.Nil(suite.T(), err)

	colReferenced, err := app.Dao().FindCollectionByNameOrId("demo2")
	assert.Nil(suite.T(), err)

	colSystem, err := app.Dao().FindCollectionByNameOrId("demo3")
	assert.Nil(suite.T(), err)
	colSystem.System = true
	assert.Nil(suite.T(), app.Dao().Save(colSystem))

	colBase, err := app.Dao().FindCollectionByNameOrId("demo1")
	assert.Nil(suite.T(), err)

	colView1, err := app.Dao().FindCollectionByNameOrId("view1")
	assert.Nil(suite.T(), err)

	colView2, err := app.Dao().FindCollectionByNameOrId("view2")
	assert.Nil(suite.T(), err)

	scenarios := []struct {
		model       *models.Collection
		expectError bool
	}{
		{colUnsaved, true},
		{colReferenced, true},
		{colSystem, true},
		{colBase, true},  // depend on view1, view2 and view2
		{colView1, true}, // view2 depend on it
		{colView2, false},
		{colView1, false}, // no longer has dependent collections
		{colBase, false},  // no longer has dependent views
		{colAuth, false},  // should delete also its related external auths
	}

	for i, s := range scenarios {
		err := app.Dao().DeleteCollection(s.model)

		hasErr := err != nil

		if hasErr != s.expectError {
			suite.T().Errorf("[%d] Expected hasErr %v, got %v (%v)", i, s.expectError, hasErr, err)
			continue
		}

		if hasErr {
			continue
		}

		if app.Dao().HasTable(s.model.Name) {
			suite.T().Errorf("[%d] Expected table/view %s to be deleted", i, s.model.Name)
		}

		// check if the external auths were deleted
		if s.model.IsAuth() {
			var total int
			err := app.Dao().ExternalAuthQuery().
				Select("count(*)").
				AndWhere(dbx.HashExp{"collectionId": s.model.Id}).
				Row(&total)

			if err != nil || total > 0 {
				suite.T().Fatalf("[%d] Expected external auths to be deleted, got %v (%v)", i, total, err)
			}
		}
	}
}

func (suite *CollectionTestSuite) TestSaveCollectionCreate() {
	app := suite.App
	collection := &models.Collection{
		Name: "new_test",
		Type: models.CollectionTypeBase,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Type: schema.FieldTypeText,
				Name: "test",
			},
		),
	}

	err := app.Dao().SaveCollection(collection)
	assert.Nil(suite.T(), err)

	assert.NotEmpty(suite.T(), collection.Id, "Expected collection id to be set")

	// check if the records table was created
	hasTable := app.Dao().HasTable(collection.Name)
	assert.True(suite.T(), hasTable, "Expected records table %s to be created", collection.Name)
	// check if the records table has the schema fields
	columns, err := app.Dao().TableColumns(collection.Name)
	assert.Nil(suite.T(), err)
	expectedColumns := []string{"id", "created", "updated", "test"}
	assert.Equal(suite.T(), len(expectedColumns), len(columns))
	for i, c := range columns {
		assert.True(suite.T(), list.ExistInSlice(c, expectedColumns), "[%d] Didn't expect record column %s", i, c)
	}
}

func (suite *CollectionTestSuite) TestSaveCollectionUpdate() {
	app := suite.App
	collection, err := app.Dao().FindCollectionByNameOrId("demo3")
	assert.Nil(suite.T(), err)

	// rename an existing schema field and add a new one
	oldField := collection.Schema.GetFieldByName("title")
	oldField.Name = "title_update"
	collection.Schema.AddField(&schema.SchemaField{
		Type: schema.FieldTypeText,
		Name: "test",
	})

	assert.Nil(suite.T(), app.Dao().SaveCollection(collection))

	// check if the records table has the schema fields
	expectedColumns := []string{"id", "created", "updated", "title_update", "test", "files"}
	columns, err := app.Dao().TableColumns(collection.Name)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(expectedColumns), len(columns))
	for i, c := range columns {
		assert.True(suite.T(), list.ExistInSlice(c, expectedColumns), "[%d] Didn't expect record column %s", i, c)
	}
}

// indirect update of a field used in view should cause view(s) update
func (suite *CollectionTestSuite) TestSaveCollectionIndirectViewsUpdate() {

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := suite.App.Dao().FindCollectionByNameOrId("demo1")
	assert.Nil(suite.T(), err)

	// update MaxSelect fields
	{
		relMany := collection.Schema.GetFieldByName("rel_many")
		relManyOpt := relMany.Options.(*schema.RelationOptions)
		relManyOpt.MaxSelect = types.Pointer(1)

		fileOne := collection.Schema.GetFieldByName("file_one")
		fileOneOpt := fileOne.Options.(*schema.FileOptions)
		fileOneOpt.MaxSelect = 10
		assert.Nil(suite.T(), suite.App.Dao().SaveCollection(collection))
	}

	// check view1 schema
	{
		view1, err := suite.App.Dao().FindCollectionByNameOrId("view1")
		assert.Nil(suite.T(), err)

		relMany := view1.Schema.GetFieldByName("rel_many")
		relManyOpt := relMany.Options.(*schema.RelationOptions)
		assert.Equal(suite.T(), 1, *relManyOpt.MaxSelect)

		fileOne := view1.Schema.GetFieldByName("file_one")
		fileOneOpt := fileOne.Options.(*schema.FileOptions)
		assert.Equal(suite.T(), 10, fileOneOpt.MaxSelect)
	}

	// check view2 schema
	{
		view2, err := suite.App.Dao().FindCollectionByNameOrId("view2")
		assert.Nil(suite.T(), err)

		relMany := view2.Schema.GetFieldByName("rel_many")
		relManyOpt := relMany.Options.(*schema.RelationOptions)
		assert.Equal(suite.T(), 1, *relManyOpt.MaxSelect)
	}
}

func (suite *CollectionTestSuite) TestSaveCollectionViewWrapping() {
	viewName := "test_wrapping"

	scenarios := []struct {
		name     string
		query    string
		expected string
	}{
		{
			"no wrapping - text field",
			"select text as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select text as id, bool from demo1)",
		},
		{
			"no wrapping - id field",
			"select text as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select text as id, bool from demo1)",
		},
		{
			"no wrapping - relation field",
			"select rel_one as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select rel_one as id, bool from demo1)",
		},
		{
			"no wrapping - select field",
			"select select_many as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select select_many as id, bool from demo1)",
		},
		{
			"no wrapping - email field",
			"select email as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select email as id, bool from demo1)",
		},
		{
			"no wrapping - datetime field",
			"select datetime as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select datetime as id, bool from demo1)",
		},
		{
			"no wrapping - url field",
			"select url as id, bool from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select url as id, bool from demo1)",
		},
		{
			"wrapping - bool field",
			"select bool as id, text as txt, url from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (SELECT cast(`id` as text) `id`,`txt`,`url` FROM (select bool as id, text as txt, url from demo1))",
		},
		{
			"wrapping - bool field (different order)",
			"select text as txt, url, bool as id from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (SELECT `txt`,`url`,cast(`id` as text) `id` FROM (select text as txt, url, bool as id from demo1))",
		},
		{
			"wrapping - json field",
			"select json as id, text, url from demo1",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (SELECT cast(`id` as text) `id`,`text`,`url` FROM (select json as id, text, url from demo1))",
		},
		{
			"wrapping - numeric id",
			"select 1 as id",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (SELECT cast(`id` as text) `id` FROM (select 1 as id))",
		},
		{
			"wrapping - expresion",
			"select ('test') as id",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (SELECT cast(`id` as text) `id` FROM (select ('test') as id))",
		},
		{
			"no wrapping - cast as text",
			"select cast('test' as text) as id",
			"CREATE VIEW `test_wrapping` AS SELECT * FROM (select cast('test' as text) as id)",
		},
	}

	for _, s := range scenarios {
		app := suite.App

		collection := &models.Collection{
			Name: viewName,
			Type: models.CollectionTypeView,
			Options: types.JsonMap{
				"query": s.query,
			},
		}

		err := app.Dao().SaveCollection(collection)
		assert.Nil(suite.T(), err)

		var sql string

		rowErr := app.Dao().DB().NewQuery("SELECT sql FROM sqlite_master WHERE type='view' AND name={:name}").
			Bind(dbx.Params{"name": viewName}).
			Row(&sql)
		assert.Nil(suite.T(), rowErr)
		assert.Equal(suite.T(), s.expected, sql)
	}
}

func (suite *CollectionTestSuite) TestImportCollections() {
	totalCollections := 11

	scenarios := []struct {
		name                   string
		jsonData               string
		deleteMissing          bool
		beforeRecordsSync      func(txDao *daos.Dao, mappedImported, mappedExisting map[string]*models.Collection) error
		expectError            bool
		expectCollectionsCount int
		beforeTestFunc         func(testApp *tests.TestApp, resultCollections []*models.Collection)
		afterTestFunc          func(testApp *tests.TestApp, resultCollections []*models.Collection)
	}{
		{
			name:                   "empty collections",
			jsonData:               `[]`,
			expectError:            true,
			expectCollectionsCount: totalCollections,
		},
		{
			name: "minimal collection import",
			jsonData: `[
				{"name": "import_test1", "schema": [{"name":"test", "type": "text"}]},
				{"name": "import_test2", "type": "auth"}
			]`,
			deleteMissing:          false,
			expectError:            false,
			expectCollectionsCount: totalCollections + 2,
		},
		{
			name: "minimal collection import + failed beforeRecordsSync",
			jsonData: `[
				{"name": "import_test", "schema": [{"name":"test", "type": "text"}]}
			]`,
			beforeRecordsSync: func(txDao *daos.Dao, mappedImported, mappedExisting map[string]*models.Collection) error {
				return errors.New("test_error")
			},
			deleteMissing:          false,
			expectError:            true,
			expectCollectionsCount: totalCollections,
		},
		{
			name: "minimal collection import + successful beforeRecordsSync",
			jsonData: `[
				{"name": "import_test", "schema": [{"name":"test", "type": "text"}]}
			]`,
			beforeRecordsSync: func(txDao *daos.Dao, mappedImported, mappedExisting map[string]*models.Collection) error {
				return nil
			},
			deleteMissing:          false,
			expectError:            false,
			expectCollectionsCount: totalCollections + 1,
		},
		{
			name: "new + update + delete system collection",
			jsonData: `[
				{
					"id":"2108348993330216960",
					"name":"demo",
					"schema":[
						{
							"id":"_2hlxbmp",
							"name":"title",
							"type":"text",
							"system":false,
							"required":true,
							"unique":false,
							"options":{
								"min":3,
								"max":null,
								"pattern":""
							}
						}
					]
				},
				{
					"name": "import1",
					"schema": [
						{
							"name":"active",
							"type":"bool"
						}
					]
				}
			]`,
			deleteMissing:          true,
			expectError:            true,
			expectCollectionsCount: totalCollections,
		},
		{
			name: "new + update + delete non-system collection",
			jsonData: `[
				{
					"id": "kpv709sk2lqbqk8",
					"system": true,
					"name": "nologin",
					"type": "auth",
					"options": {
						"allowEmailAuth": false,
						"allowOAuth2Auth": false,
						"allowUsernameAuth": false,
						"exceptEmailDomains": [],
						"manageRule": "@request.auth.collectionName = 'users'",
						"minPasswordLength": 8,
						"onlyEmailDomains": [],
						"requireEmail": true
					},
					"listRule": "",
					"viewRule": "",
					"createRule": "",
					"updateRule": "",
					"deleteRule": "",
					"schema": [
						{
							"id": "x8zzktwe",
							"name": "name",
							"type": "text",
							"system": false,
							"required": false,
							"unique": false,
							"options": {
								"min": null,
								"max": null,
								"pattern": ""
							}
						}
					]
				},
				{
					"id":"2108348993330216960",
					"name":"demo1_rename",
					"schema":[
						{
							"id":"_2hlxbmp",
							"name":"title",
							"type":"text",
							"system":false,
							"required":true,
							"unique":false,
							"options":{
								"min":3,
								"max":null,
								"pattern":""
							}
						}
					]
				},
				{
					"id": "test_deleted_collection_name_reuse",
					"name": "demo2",
					"schema": [
						{
							"id":"fz6iql2m",
							"name":"active",
							"type":"bool"
						}
					]
				},
				{
					"id": "test_new_view",
					"name": "new_view",
					"type": "view",
					"options": {
						"query": "select id from demo2"
					}
				}
			]`,
			deleteMissing:          true,
			expectError:            false,
			expectCollectionsCount: 4,
		},
		{
			name: "test with deleteMissing: false",
			jsonData: `[
				{
					"id":"2108348993330216960",
					"name":"demo1",
					"schema":[
						{
							"id":"_2hlxbmp",
							"name":"title",
							"type":"text",
							"system":false,
							"required":true,
							"unique":false,
							"options":{
								"min":3,
								"max":null,
								"pattern":""
							}
						},
						{
							"id":"_2hlxbmp",
							"name":"field_with_duplicate_id",
							"type":"text",
							"system":false,
							"required":true,
							"unique":false,
							"options":{
								"min":3,
								"max":null,
								"pattern":""
							}
						},
						{
							"id":"abcd_import",
							"name":"new_field",
							"type":"text"
						}
					]
				},
				{
					"name": "new_import",
					"schema": [
						{
							"id":"abcd_import",
							"name":"active",
							"type":"bool"
						}
					]
				}
			]`,
			deleteMissing:          false,
			expectError:            false,
			expectCollectionsCount: totalCollections + 1,
			afterTestFunc: func(testApp *tests.TestApp, resultCollections []*models.Collection) {
				expectedCollectionFields := map[string]int{
					"nologin":    1,
					"demo1":      15,
					"demo2":      2,
					"demo3":      2,
					"demo4":      11,
					"demo5":      6,
					"new_import": 1,
				}
				for name, expectedCount := range expectedCollectionFields {
					collection, err := testApp.Dao().FindCollectionByNameOrId(name)
					assert.Nil(suite.T(), err)
					assert.Equal(suite.T(), expectedCount, len(collection.Schema.Fields()))
				}
			},
		},
	}

	for _, scenario := range scenarios {
		testApp := suite.App

		importedCollections := []*models.Collection{}

		// load data
		loadErr := json.Unmarshal([]byte(scenario.jsonData), &importedCollections)
		assert.Nil(suite.T(), loadErr)

		err := testApp.Dao().ImportCollections(importedCollections, scenario.deleteMissing, scenario.beforeRecordsSync)

		hasErr := err != nil
		assert.Equal(suite.T(), scenario.expectError, hasErr)

		// check collections count
		collections := []*models.Collection{}
		assert.Nil(suite.T(), testApp.Dao().CollectionQuery().All(&collections))
		assert.Equal(suite.T(), scenario.expectCollectionsCount, len(collections))

		if scenario.afterTestFunc != nil {
			scenario.afterTestFunc(testApp, collections)
		}
	}
}

type CollectionTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (suite *CollectionTestSuite) SetupTest() {
	app, _ := tests.NewTestApp()
	suite.Var = 5
	suite.App = app
}

func (suite *CollectionTestSuite) TearDownTest() {
	suite.App.Cleanup()
}

func (suite *CollectionTestSuite) SetupSuite() {
	fmt.Println("setup suite")
}

func (suite *CollectionTestSuite) TearDownSuite() {
	fmt.Println("teardown suite")
}

func TestCollectionTestSuite(t *testing.T) {
	suite.Run(t, new(CollectionTestSuite))
}
