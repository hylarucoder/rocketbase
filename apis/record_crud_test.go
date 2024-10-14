package apis_test

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
)

func (suite *RecordCrudTestSuite) TestRecordCrudList() {
	t := suite.T()
	scenarios := []tests.ApiScenario{
		{
			Name:            "missing collection",
			Method:          http.MethodGet,
			Url:             "/api/collections/missing/records",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "unauthenticated trying to access nil rule collection (aka. need admin auth)",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo1/records",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record trying to access nil rule collection (aka. need admin auth)",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "public collection but with admin only filter param (aka. @collection, @request, etc.)",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo2/records?filter=%40collection.demo2.title='test1'",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "public collection but with admin only sort param (aka. @collection, @request, etc.)",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo2/records?sort=@request.auth.title",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "public collection but with ENCODED admin only filter/sort (aka. @collection)",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo2/records?filter=%40collection.demo2.title%3D%27test1%27",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection",
			Method:         http.MethodGet,
			Url:            "/api/collections/demo2/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3479948460419978246"`,
				`"id":"3479948460512252935"`,
				`"id":"3479948460562584584"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection (using the collection id)",
			Method:         http.MethodGet,
			Url:            "/api/collections/2108349190391201792/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3479948460419978246"`,
				`"id":"3479948460512252935"`,
				`"id":"3479948460562584584"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin trying to access nil rule collection (aka. need admin auth)",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3479947686654776325"`,
				`"id":"3479947686587667460"`,
				`"id":"3479947686461838339"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "valid query params",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records?filter=text~'test'&sort=-bool",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":2`,
				`"items":[{`,
				//`"id":"3479947686654776325"`,
				`"id":"3479947686461838339"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "invalid filter",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records?filter=invalid~'test'",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			// TODO: admin fix
			Name:   "expand relations",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records?expand=rel_one,rel_many.rel,missing&perPage=2&sort=created",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":2`,
				`"totalPages":2`,
				`"totalItems":3`,
				`"items":[{`,
				`"collectionName":"demo1"`,
				`"id":"3479947686461838339"`,
				`"id":"3479946329126343681"`,
				`"expand":{`,
				`"rel_one":""`,
				`"rel_one":{"`,
				`"rel_many":[{`,
				`"rel":{`,
				`"rel":""`,
				`"json":[1,2,3]`,
				`"select_many":["optionB","optionC"]`,
				`"select_many":["optionB"]`,
				// subrel items
				`"id":"3479948460419978246"`,
				`"id":"3479948460512252935"`,
				// email visibility should be ignored for admins even in expanded rels
				`"email":"test@example.com"`,
				`"email":"test2@example.com"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record model that DOESN'T match the collection list rule",
			Method: http.MethodGet,
			Url:    "/api/collections/demo3/records",
			RequestHeaders: map[string]string{
				// TODO: confirm users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":0`,
				`"items":[]`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that matches the collection list rule",
			Method: http.MethodGet,
			Url:    "/api/collections/demo3/records",
			RequestHeaders: map[string]string{
				// TODO: clients, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":4`,
				`"items":[{`,
				`"id":"3479958939318096910"`,
				`"id":"3479958939469091857"`,
				`"id":"3479958939418760208"`,
				`"id":"3479958939368428559"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   ":rule modifer",
			Method: http.MethodGet,
			Url:    "/api/collections/demo5/records",
			// TODO: fix 400?
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":1`,
				`"items":[{`,
				`"id":"3479953184028365833"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "multi-match - at least one of",
			Method:         http.MethodGet,
			Url:            "/api/collections/demo4/records?filter=" + url.QueryEscape("rel_many_no_cascade_required.files:length?=2"),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":1`,
				`"items":[{`,
				`"id":"qzaqccwrmva4o1n"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "multi-match - all",
			Method:         http.MethodGet,
			Url:            "/api/collections/demo4/records?filter=" + url.QueryEscape("rel_many_no_cascade_required.files:length=2"),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":0`,
				`"totalItems":0`,
				`"items":[]`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// auth collection
		// -----------------------------------------------------------
		{
			Name:           "check email visibility as guest",
			Method:         http.MethodGet,
			Url:            "/api/collections/nologin/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3480271880273794066"`,
				`"id":"3480271880340902931"`,
				`"id":"3480271880374457364"`,
				`"email":"test2@example.com"`,
				`"emailVisibility":true`,
				`"emailVisibility":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
				`"email":"test@example.com"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as any authenticated record",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records",
			RequestHeaders: map[string]string{
				// clients, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3480271880273794066"`,
				`"id":"3480271880340902931"`,
				`"id":"3480271880374457364"`,
				`"email":"test2@example.com"`,
				`"emailVisibility":true`,
				`"emailVisibility":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
				`"email":"test@example.com"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as manage auth record",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3480271880273794066"`,
				`"id":"3480271880340902931"`,
				`"id":"3480271880374457364"`,
				`"email":"test@example.com"`,
				`"email":"test2@example.com"`,
				`"email":"test3@example.com"`,
				`"emailVisibility":true`,
				`"emailVisibility":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as admin",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records",
			RequestHeaders: map[string]string{
				// admin, test@example.com
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3480271880273794066"`,
				`"id":"3480271880340902931"`,
				`"id":"3480271880374457364"`,
				`"email":"test@example.com"`,
				`"email":"test2@example.com"`,
				`"email":"test3@example.com"`,
				`"emailVisibility":true`,
				`"emailVisibility":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check self email visibility resolver",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records",
			RequestHeaders: map[string]string{
				// nologin, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"3480271880273794066"`,
				`"id":"3480271880340902931"`,
				`"id":"3480271880374457364"`,
				`"email":"test2@example.com"`,
				`"email":"test@example.com"`,
				`"emailVisibility":true`,
				`"emailVisibility":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// view collection
		// -----------------------------------------------------------
		{
			Name:           "public view records",
			Method:         http.MethodGet,
			Url:            "/api/collections/view2/records?filter=state=\"false\"",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":2`,
				`"items":[{`,
				`"id":"3479947686654776325"`,
				`"id":"3479947686587667460"`,
			},
			NotExpectedContent: []string{
				`"created"`,
				`"updated"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "guest that doesn't match the view collection list rule",
			Method:         http.MethodGet,
			Url:            "/api/collections/view1/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":0`,
				`"totalItems":0`,
				`"items":[]`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that matches the view collection list rule",
			Method: http.MethodGet,
			Url:    "/api/collections/view1/records",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":1`,
				`"items":[{`,
				`"id":"3479947686461838339"`,
				`"bool":true`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "view collection with numeric ids",
			Method:         http.MethodGet,
			Url:            "/api/collections/numeric_id_view/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalPages":1`,
				`"totalItems":2`,
				`"items":[{`,
				`"id":"1"`,
				`"id":"2"`,
			},
			ExpectedEvents: map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *RecordCrudTestSuite) TestRecordCrudView() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "missing collection",
			Method:          http.MethodGet,
			Url:             "/api/collections/missing/records/3479948460419978246",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "missing record",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo2/records/missing",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "unauthenticated trying to access nil rule collection (aka. need admin auth)",
			Method:          http.MethodGet,
			Url:             "/api/collections/demo1/records/3479947686587667460",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record trying to access nil rule collection (aka. need admin auth)",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records/3479947686587667460",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that doesn't match the collection view rule",
			Method: http.MethodGet,
			Url:    "/api/collections/users/records/bgs820n361vj1qd",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection view",
			Method:         http.MethodGet,
			Url:            "/api/collections/demo2/records/3479948460419978246",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479948460419978246"`,
				`"collectionName":"demo2"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection view (using the collection id)",
			Method:         http.MethodGet,
			Url:            "/api/collections/2108349190391201792/records/3479948460419978246",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479948460419978246"`,
				`"collectionName":"demo2"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin trying to access nil rule collection view (aka. need admin auth)",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records/3479947686587667460",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479947686587667460"`,
				`"collectionName":"demo1"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that does match the collection view rule",
			Method: http.MethodGet,
			Url:    "/api/collections/users/records/2107977397063122944",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"2107977397063122944"`,
				`"collectionName":"users"`,
				// owners can always view their email
				`"emailVisibility":false`,
				`"email":"test@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "expand relations",
			Method: http.MethodGet,
			Url:    "/api/collections/demo1/records/3479947686461838339?expand=rel_one,rel_many.rel,missing&perPage=2&sort=created",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479947686461838339"`,
				`"collectionName":"demo1"`,
				`"rel_many":[{`,
				`"rel_one":{`,
				`"collectionName":"users"`,
				`"id":"3479946329126343681"`,
				`"expand":{"rel":{`,
				`"id":"3479947686587667460"`,
				`"collectionName":"demo2"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// auth collection
		// -----------------------------------------------------------
		{
			Name:           "check email visibility as guest",
			Method:         http.MethodGet,
			Url:            "/api/collections/nologin/records/3480271880374457364",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880374457364"`,
				`"emailVisibility":false`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as any authenticated record",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records/3480271880374457364",
			RequestHeaders: map[string]string{
				// clients, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880374457364"`,
				`"emailVisibility":false`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
				`"email":"test3@example.com"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as manage auth record",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records/3480271880374457364",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880374457364"`,
				`"emailVisibility":false`,
				`"email":"test3@example.com"`,
				`"verified":true`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check email visibility as admin",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records/3480271880374457364",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880374457364"`,
				`"emailVisibility":false`,
				`"email":"test3@example.com"`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "check self email visibility resolver",
			Method: http.MethodGet,
			Url:    "/api/collections/nologin/records/3480271880273794066",
			RequestHeaders: map[string]string{
				// nologin, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880273794066"`,
				`"email":"test@example.com"`,
				`"emailVisibility":false`,
				`"verified":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// view collection
		// -----------------------------------------------------------
		{
			Name:           "public view record",
			Method:         http.MethodGet,
			Url:            "/api/collections/view2/records/3479947686461838339",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479947686461838339"`,
				`"state":true`,
				`"file_many":["`,
				`"rel_many":["`,
			},
			NotExpectedContent: []string{
				`"created"`,
				`"updated"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "guest that doesn't match the view collection view rule",
			Method:          http.MethodGet,
			Url:             "/api/collections/view1/records/3479947686461838339",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authenticated record that matches the view collection view rule",
			Method: http.MethodGet,
			Url:    "/api/collections/view1/records/3479947686461838339",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479947686461838339"`,
				`"bool":true`,
				`"text":"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "view record with numeric id",
			Method:         http.MethodGet,
			Url:            "/api/collections/numeric_id_view/records/1",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"1"`,
			},
			ExpectedEvents: map[string]int{"OnRecordViewRequest": 1},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *RecordCrudTestSuite) TestRecordCrudDelete() {
	t := suite.T()

	ensureDeletedFiles := func(app *tests.TestApp, collectionId string, recordId string) {
		storageDir := filepath.Join(app.DataDir(), "storage", collectionId, recordId)

		entries, _ := os.ReadDir(storageDir)
		if len(entries) != 0 {
			t.Errorf("Expected empty/deleted dir, found %d", len(entries))
		}
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "missing collection",
			Method:          http.MethodDelete,
			Url:             "/api/collections/missing/records/3479948460419978246",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "missing record",
			Method:          http.MethodDelete,
			Url:             "/api/collections/demo2/records/missing",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "unauthenticated trying to delete nil rule collection (aka. need admin auth)",
			Method:          http.MethodDelete,
			Url:             "/api/collections/demo1/records/3479947686587667460",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record trying to delete nil rule collection (aka. need admin auth)",
			Method: http.MethodDelete,
			Url:    "/api/collections/demo1/records/3479947686587667460",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that doesn't match the collection delete rule",
			Method: http.MethodDelete,
			Url:    "/api/collections/users/records/bgs820n361vj1qd",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "trying to delete a view collection record",
			Method:          http.MethodDelete,
			Url:             "/api/collections/view1/records/3479947686587667460",
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection record delete",
			Method:         http.MethodDelete,
			Url:            "/api/collections/nologin/records/3480271880273794066",
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          1,
				"OnModelBeforeDelete":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "public collection record delete (using the collection id as identifier)",
			Method:         http.MethodDelete,
			Url:            "/api/collections/2108654300501639168/records/3480271880273794066",
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          1,
				"OnModelBeforeDelete":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin trying to delete nil rule collection view (aka. need admin auth)",
			Method: http.MethodDelete,
			Url:    "/api/collections/clients/records/3479946329210229762",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          1,
				"OnModelBeforeDelete":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "OnRecordAfterDeleteRequest error response",
			Method: http.MethodDelete,
			Url:    "/api/collections/clients/records/3479946329210229762",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          1,
				"OnModelBeforeDelete":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authenticated record that match the collection delete rule",
			Method: http.MethodDelete,
			Url:    "/api/collections/users/records/2107977397063122944",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			Delay:          100 * time.Millisecond,
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          3, // +2 because of the external auths
				"OnModelBeforeDelete":         3, // +2 because of the external auths
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				ensureDeletedFiles(app, "_pb_users_auth_", "2107977397063122944")

				// check if all the external auths records were deleted
				collection, _ := app.Dao().FindCollectionByNameOrId("users")
				record := models.NewRecord(collection)
				record.Id = "2107977397063122944"
				externalAuths, err := app.Dao().FindAllExternalAuthsByRecord(record)
				if err != nil {
					t.Errorf("Failed to fetch external auths: %v", err)
				}
				if len(externalAuths) > 0 {
					t.Errorf("Expected the linked external auths to be deleted, got %d", len(externalAuths))
				}
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "@request :isset (rule failure check)",
			Method:          http.MethodDelete,
			Url:             "/api/collections/demo5/records/la4y2w4o98acwuj",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:           "@request :isset (rule pass check)",
			Method:         http.MethodDelete,
			Url:            "/api/collections/demo5/records/la4y2w4o98acwuj?test=1",
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelAfterDelete":          1,
				"OnModelBeforeDelete":         1,
				"OnRecordAfterDeleteRequest":  1,
				"OnRecordBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// cascade delete checks
		// -----------------------------------------------------------
		{
			Name:   "trying to delete a record while being part of a non-cascade required relation",
			Method: http.MethodDelete,
			Url:    "/api/collections/demo3/records/7nwo8tuiatetxdm",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeDeleteRequest": 1,
				"OnModelBeforeUpdate":         2, // self_rel_many update of test1 record + rel_one_cascade demo4 cascaded in demo5
				"OnModelBeforeDelete":         2, // the record itself + rel_one_cascade of test1 record
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "delete a record with non-cascade references",
			Method: http.MethodDelete,
			Url:    "/api/collections/demo3/records/1tmknxy2868d869",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelBeforeDelete":         1,
				"OnModelAfterDelete":          1,
				"OnModelBeforeUpdate":         2,
				"OnModelAfterUpdate":          2,
				"OnRecordBeforeDeleteRequest": 1,
				"OnRecordAfterDeleteRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "delete a record with cascade references",
			Method: http.MethodDelete,
			Url:    "/api/collections/users/records/2108356222582259712",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			Delay:          100 * time.Millisecond,
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelBeforeDelete":         2,
				"OnModelAfterDelete":          2,
				"OnModelBeforeUpdate":         2,
				"OnModelAfterUpdate":          2,
				"OnRecordBeforeDeleteRequest": 1,
				"OnRecordAfterDeleteRequest":  1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				recId := "3479947686461838339"
				rec, _ := app.Dao().FindRecordById("demo1", recId, nil)
				if rec != nil {
					t.Errorf("Expected record %s to be cascade deleted", recId)
				}
				ensureDeletedFiles(app, "2108348993330216960", recId)
				ensureDeletedFiles(app, "_pb_users_auth_", "2108356222582259712")
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *RecordCrudTestSuite) TestRecordCrudCreate() {
	t := suite.T()

	formData, mp, err := tests.MockMultipartData(map[string]string{
		"title": "title_test",
	}, "files")
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "missing collection",
			Method:          http.MethodPost,
			Url:             "/api/collections/missing/records",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "guest trying to access nil-rule collection",
			Method:          http.MethodPost,
			Url:             "/api/collections/demo1/records",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record trying to access nil-rule collection",
			Method: http.MethodPost,
			Url:    "/api/collections/demo1/records",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "trying to create a new view collection record",
			Method:          http.MethodPost,
			Url:             "/api/collections/view1/records",
			Body:            strings.NewReader(`{"text":"new"}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "submit nil body",
			Method:          http.MethodPost,
			Url:             "/api/collections/demo2/records",
			Body:            nil,
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "submit invalid format",
			Method:          http.MethodPost,
			Url:             "/api/collections/demo2/records",
			Body:            strings.NewReader(`{"`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "submit empty json body",
			Method:         http.MethodPost,
			Url:            "/api/collections/nologin/records",
			Body:           strings.NewReader(`{}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"email":{"code":"validation_required"`,
				`"password":{"code":"validation_required"`,
				`"passwordConfirm":{"code":"validation_required"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "guest submit in public collection",
			Method:         http.MethodPost,
			Url:            "/api/collections/demo2/records",
			Body:           strings.NewReader(`{"title":"new"}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":`,
				`"title":"new"`,
				`"active":false`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "guest trying to submit in restricted collection",
			Method:          http.MethodPost,
			Url:             "/api/collections/demo3/records",
			Body:            strings.NewReader(`{"title":"test123"}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record submit in restricted collection (rule failure check)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body:   strings.NewReader(`{"title":"test123"}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record submit in restricted collection (rule pass check) + expand relations",
			Method: http.MethodPost,
			Url:    "/api/collections/demo4/records?expand=missing,rel_one_no_cascade,rel_many_no_cascade_required",
			Body: strings.NewReader(`{
				"title":"test123",
				"rel_one_no_cascade":"3479958939318096910",
				"rel_one_no_cascade_required":"7nwo8tuiatetxdm",
				"rel_one_cascade":"3479958939318096910",
				"rel_many_no_cascade":"3479958939318096910",
				"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"],
				"rel_many_cascade":"lcl9d87w22ml6jy"
			}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":`,
				`"title":"test123"`,
				`"rel_one_no_cascade":"3479958939318096910"`,
				`"rel_one_no_cascade_required":"7nwo8tuiatetxdm"`,
				`"rel_one_cascade":"3479958939318096910"`,
				`"rel_many_no_cascade":["3479958939318096910"]`,
				`"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"]`,
				`"rel_many_cascade":["lcl9d87w22ml6jy"]`,
			},
			NotExpectedContent: []string{
				// the users auth records don't have access to view the demo3 expands
				`"expand":{`,
				`"missing"`,
				`"id":"3479958939318096910"`,
				`"id":"7nwo8tuiatetxdm"`,
				`"id":"lcl9d87w22ml6jy"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "admin submit in restricted collection (rule skip check) + expand relations",
			Method: http.MethodPost,
			Url:    "/api/collections/demo4/records?expand=missing,rel_one_no_cascade,rel_many_no_cascade_required",
			Body: strings.NewReader(`{
				"title":"test123",
				"rel_one_no_cascade":"3479958939318096910",
				"rel_one_no_cascade_required":"7nwo8tuiatetxdm",
				"rel_one_cascade":"3479958939318096910",
				"rel_many_no_cascade":"3479958939318096910",
				"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"],
				"rel_many_cascade":"lcl9d87w22ml6jy"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":`,
				`"title":"test123"`,
				`"rel_one_no_cascade":"3479958939318096910"`,
				`"rel_one_no_cascade_required":"7nwo8tuiatetxdm"`,
				`"rel_one_cascade":"3479958939318096910"`,
				`"rel_many_no_cascade":["3479958939318096910"]`,
				`"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"]`,
				`"rel_many_cascade":["lcl9d87w22ml6jy"]`,
				`"expand":{`,
				`"id":"3479958939318096910"`,
				`"id":"7nwo8tuiatetxdm"`,
				`"id":"lcl9d87w22ml6jy"`,
			},
			NotExpectedContent: []string{
				`"missing"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "submit via multipart form data",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body:   formData,
			RequestHeaders: map[string]string{
				"Content-Type":  mp.FormDataContentType(),
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"`,
				`"title":"title_test"`,
				`"files":["`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "unique field error check",
			Method: http.MethodPost,
			Url:    "/api/collections/demo2/records",
			Body: strings.NewReader(`{
				"title":"test2"
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"title":{`,
				`"code":"validation_not_unique"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "OnRecordAfterCreateRequest error response",
			Method: http.MethodPost,
			Url:    "/api/collections/demo2/records",
			Body:   strings.NewReader(`{"title":"new"}`),
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// ID checks
		// -----------------------------------------------------------
		{
			Name:   "invalid custom insertion id (less than 15 chars)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body: strings.NewReader(`{
				"id": "12345678901234",
				"title": "test"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"id":{"code":"validation_length_invalid"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "invalid custom insertion id (more than 15 chars)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body: strings.NewReader(`{
				"id": "1234567890123456",
				"title": "test"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"id":{"code":"validation_length_invalid"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "valid custom insertion id (exactly 15 chars)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body: strings.NewReader(`{
				"id": "123456789012345",
				"title": "test"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"123456789012345"`,
				`"title":"test"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "valid custom insertion id existing in another non-auth collection",
			Method: http.MethodPost,
			Url:    "/api/collections/demo3/records",
			Body: strings.NewReader(`{
				"id": "3479948460419978246",
				"title": "test"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479948460419978246"`,
				`"title":"test"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
				"OnRecordAfterCreateRequest":  1,
				"OnModelBeforeCreate":         1,
				"OnModelAfterCreate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "valid custom insertion auth id duplicating in another auth collection",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			Body: strings.NewReader(`{
				"id":"3479946329210229762",
				"title":"test",
				"password":"1234567890",
				"passwordConfirm":"1234567890"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeCreateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// fields modifier checks
		// -----------------------------------------------------------
		{
			Name:   "trying to delete a record while being part of a non-cascade required relation",
			Method: http.MethodDelete,
			Url:    "/api/collections/demo3/records/7nwo8tuiatetxdm",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeDeleteRequest": 1,
				"OnModelBeforeUpdate":         2, // self_rel_many update of test1 record + rel_one_cascade demo4 cascaded in demo5
				"OnModelBeforeDelete":         2, // the record itself + rel_one_cascade of test1 record
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// check whether if @request.data modifer fields are properly resolved
		// -----------------------------------------------------------
		{
			Name:   "@request.data.field with compute modifers (rule failure check)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo5/records",
			Body: strings.NewReader(`{
				"total":1,
				"total+":4,
				"total-":1
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "@request.data.field with compute modifers (rule pass check)",
			Method: http.MethodPost,
			Url:    "/api/collections/demo5/records",
			Body: strings.NewReader(`{
				"total":1,
				"total+":3,
				"total-":1
			}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"`,
				`"collectionName":"demo5"`,
				`"total":3`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterCreate":          1,
				"OnModelBeforeCreate":         1,
				"OnRecordAfterCreateRequest":  1,
				"OnRecordBeforeCreateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// auth records
		// -----------------------------------------------------------
		{
			Name:   "auth record with invalid data",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			Body: strings.NewReader(`{
				"id":"o1y0pd786mq",
				"username":"Users75657",
				"email":"invalid",
				"password":"1234567",
				"passwordConfirm":"1234560"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"id":{"code":"validation_length_invalid"`,
				`"username":{"code":"validation_invalid_username"`, // for duplicated case-insensitive username
				`"email":{"code":"validation_is_email"`,
				`"password":{"code":"validation_length_out_of_range"`,
				`"passwordConfirm":{"code":"validation_values_mismatch"`,
			},
			NotExpectedContent: []string{
				// schema fields are not checked if the base fields has errors
				`"rel":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid base fields but invalid schema data",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"rel":"invalid"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"rel":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid data and explicitly verified state by guest",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"verified":true
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"verified":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid data and explicitly verified state by random user",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"emailVisibility":true,
				"verified":true
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"verified":{"code":`,
			},
			NotExpectedContent: []string{
				`"emailVisibility":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid data by admin",
			Method: http.MethodPost,
			Url:    "/api/collections/users/records",
			Body: strings.NewReader(`{
				"id":"o1o1y0pd78686mq",
				"username":"test.valid",
				"email":"new@example.com",
				"password":"12345678",
				"passwordConfirm":"12345678",
				"rel":"achvryl401bhse3",
				"emailVisibility":true,
				"verified":true
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"o1o1y0pd78686mq"`,
				`"username":"test.valid"`,
				`"email":"new@example.com"`,
				`"rel":"achvryl401bhse3"`,
				`"emailVisibility":true`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterCreate":          1,
				"OnModelBeforeCreate":         1,
				"OnRecordAfterCreateRequest":  1,
				"OnRecordBeforeCreateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid data by auth record with manage access",
			Method: http.MethodPost,
			Url:    "/api/collections/nologin/records",
			Body: strings.NewReader(`{
				"email":"new@example.com",
				"password":"12345678",
				"passwordConfirm":"12345678",
				"name":"test_name",
				"emailVisibility":true,
				"verified":true
			}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"`,
				`"username":"`,
				`"email":"new@example.com"`,
				`"name":"test_name"`,
				`"emailVisibility":true`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterCreate":          1,
				"OnModelBeforeCreate":         1,
				"OnRecordAfterCreateRequest":  1,
				"OnRecordBeforeCreateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *RecordCrudTestSuite) TestRecordCrudUpdate() {
	t := suite.T()

	formData, mp, err := tests.MockMultipartData(map[string]string{
		"title": "title_test",
	}, "files")
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "missing collection",
			Method:          http.MethodPatch,
			Url:             "/api/collections/missing/records/3479948460419978246",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "guest trying to access nil-rule collection record",
			Method:          http.MethodPatch,
			Url:             "/api/collections/demo1/records/3479947686587667460",
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record trying to access nil-rule collection",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo1/records/3479947686587667460",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "submit invalid body",
			Method:          http.MethodPatch,
			Url:             "/api/collections/demo2/records/34799484604199782460",
			Body:            strings.NewReader(`{"`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "trying to update a view collection record",
			Method:          http.MethodPatch,
			Url:             "/api/collections/view1/records/3479947686654776325",
			Body:            strings.NewReader(`{"text":"new"}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "submit nil body",
			Method:          http.MethodPatch,
			Url:             "/api/collections/demo2/records/3479948460419978246",
			Body:            nil,
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "submit empty body (aka. no fields change)",
			Method:         http.MethodPatch,
			Url:            "/api/collections/demo2/records/3479948460419978246",
			Body:           strings.NewReader(`{}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"collectionName":"demo2"`,
				`"id":"3479948460419978246"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "trigger field validation",
			Method:         http.MethodPatch,
			Url:            "/api/collections/demo2/records/3479948460419978246",
			Body:           strings.NewReader(`{"title":"a"}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`data":{`,
				`"title":{"code":"validation_min_text_constraint"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:           "guest submit in public collection",
			Method:         http.MethodPatch,
			Url:            "/api/collections/demo2/records/3479948460419978246",
			Body:           strings.NewReader(`{"title":"new"}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479948460419978246"`,
				`"title":"new"`,
				`"active":true`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnRecordAfterUpdateRequest":  1,
				"OnModelBeforeUpdate":         1,
				"OnModelAfterUpdate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:            "guest trying to submit in restricted collection",
			Method:          http.MethodPatch,
			Url:             "/api/collections/demo3/records/3479958939318096910",
			Body:            strings.NewReader(`{"title":"new"}`),
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record submit in restricted collection (rule failure check)",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo3/records/3479958939318096910",
			Body:   strings.NewReader(`{"title":"new"}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record submit in restricted collection (rule pass check) + expand relations",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo4/records/i9naidtvr6qsgb4?expand=missing,rel_one_no_cascade,rel_many_no_cascade_required",
			Body: strings.NewReader(`{
				"title":"test123",
				"rel_one_no_cascade":"3479958939318096910",
				"rel_one_no_cascade_required":"7nwo8tuiatetxdm",
				"rel_one_cascade":"3479958939318096910",
				"rel_many_no_cascade":"3479958939318096910",
				"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"],
				"rel_many_cascade":"lcl9d87w22ml6jy"
			}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"i9naidtvr6qsgb4"`,
				`"title":"test123"`,
				`"rel_one_no_cascade":"3479958939318096910"`,
				`"rel_one_no_cascade_required":"7nwo8tuiatetxdm"`,
				`"rel_one_cascade":"3479958939318096910"`,
				`"rel_many_no_cascade":["3479958939318096910"]`,
				`"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"]`,
				`"rel_many_cascade":["lcl9d87w22ml6jy"]`,
			},
			NotExpectedContent: []string{
				// the users auth records don't have access to view the demo3 expands
				`"expand":{`,
				`"missing"`,
				`"id":"3479958939318096910"`,
				`"id":"7nwo8tuiatetxdm"`,
				`"id":"lcl9d87w22ml6jy"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnRecordAfterUpdateRequest":  1,
				"OnModelBeforeUpdate":         1,
				"OnModelAfterUpdate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "admin submit in restricted collection (rule skip check) + expand relations",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo4/records/i9naidtvr6qsgb4?expand=missing,rel_one_no_cascade,rel_many_no_cascade_required",
			Body: strings.NewReader(`{
				"title":"test123",
				"rel_one_no_cascade":"3479958939318096910",
				"rel_one_no_cascade_required":"7nwo8tuiatetxdm",
				"rel_one_cascade":"3479958939318096910",
				"rel_many_no_cascade":"3479958939318096910",
				"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"],
				"rel_many_cascade":"lcl9d87w22ml6jy"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"i9naidtvr6qsgb4"`,
				`"title":"test123"`,
				`"rel_one_no_cascade":"3479958939318096910"`,
				`"rel_one_no_cascade_required":"7nwo8tuiatetxdm"`,
				`"rel_one_cascade":"3479958939318096910"`,
				`"rel_many_no_cascade":["3479958939318096910"]`,
				`"rel_many_no_cascade_required":["7nwo8tuiatetxdm","lcl9d87w22ml6jy"]`,
				`"rel_many_cascade":["lcl9d87w22ml6jy"]`,
				`"expand":{`,
				`"id":"3479958939318096910"`,
				`"id":"7nwo8tuiatetxdm"`,
				`"id":"lcl9d87w22ml6jy"`,
			},
			NotExpectedContent: []string{
				`"missing"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnRecordAfterUpdateRequest":  1,
				"OnModelBeforeUpdate":         1,
				"OnModelAfterUpdate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "submit via multipart form data",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo3/records/3479958939318096910",
			Body:   formData,
			RequestHeaders: map[string]string{
				"Content-Type":  mp.FormDataContentType(),
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3479958939318096910"`,
				`"title":"title_test"`,
				`"files":["`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnRecordAfterUpdateRequest":  1,
				"OnModelBeforeUpdate":         1,
				"OnModelAfterUpdate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "OnRecordAfterUpdateRequest error response",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo2/records/3479948460419978246",
			Body:   strings.NewReader(`{"title":"new"}`),
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnRecordAfterUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnRecordAfterUpdateRequest":  1,
				"OnModelBeforeUpdate":         1,
				"OnModelAfterUpdate":          1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "try to change the id of an existing record",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo3/records/3479958939318096910",
			Body: strings.NewReader(`{
				"id": "mk5fmymtx4wspra"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"id":{"code":"validation_in_invalid"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "unique field error check",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo2/records/3479948460512252935",
			Body: strings.NewReader(`{
				"title":"test2"
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"title":{`,
				`"code":"validation_not_unique"`,
			},
			ExpectedEvents: map[string]int{
				"OnRecordBeforeUpdateRequest": 1,
				"OnModelBeforeUpdate":         1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// check whether if @request.data modifer fields are properly resolved
		// -----------------------------------------------------------
		{
			Name:   "@request.data.field with compute modifers (rule failure check)",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo5/records/la4y2w4o98acwuj",
			Body: strings.NewReader(`{
				"total+":3,
				"total-":1
			}`),
			ExpectedStatus: 404,
			ExpectedContent: []string{
				`"data":{}`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "@request.data.field with compute modifers (rule pass check)",
			Method: http.MethodPatch,
			Url:    "/api/collections/demo5/records/la4y2w4o98acwuj",
			Body: strings.NewReader(`{
				"total+":2,
				"total-":1
			}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"la4y2w4o98acwuj"`,
				`"collectionName":"demo5"`,
				`"total":3`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},

		// auth records
		// -----------------------------------------------------------
		{
			Name:   "auth record with invalid data",
			Method: http.MethodPatch,
			Url:    "/api/collections/users/records/bgs820n361vj1qd",
			Body: strings.NewReader(`{
				"username":"Users75657",
				"email":"invalid",
				"password":"1234567",
				"passwordConfirm":"1234560",
				"verified":false
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"username":{"code":"validation_invalid_username"`, // for duplicated case-insensitive username
				`"email":{"code":"validation_is_email"`,
				`"password":{"code":"validation_length_out_of_range"`,
				`"passwordConfirm":{"code":"validation_values_mismatch"`,
			},
			NotExpectedContent: []string{
				// admins are allowed to change the verified state
				`"verified"`,
				// schema fields are not checked if the base fields has errors
				`"rel":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "auth record with valid base fields but invalid schema data",
			Method: http.MethodPatch,
			Url:    "/api/collections/users/records/bgs820n361vj1qd",
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"rel":"invalid"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"rel":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "try to change account managing fields by guest",
			Method: http.MethodPatch,
			Url:    "/api/collections/nologin/records/phhq3wr65cap535",
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"emailVisibility":true,
				"verified":true
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"verified":{"code":`,
				`"oldPassword":{"code":`,
			},
			NotExpectedContent: []string{
				`"emailVisibility":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "try to change account managing fields by auth record (owner)",
			Method: http.MethodPatch,
			Url:    "/api/collections/users/records/2107977397063122944",
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			Body: strings.NewReader(`{
				"password":"12345678",
				"passwordConfirm":"12345678",
				"emailVisibility":true,
				"verified":true
			}`),
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"verified":{"code":`,
				`"oldPassword":{"code":`,
			},
			NotExpectedContent: []string{
				`"emailVisibility":{"code":`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "try to change account managing fields by auth record with managing rights",
			Method: http.MethodPatch,
			Url:    "/api/collections/nologin/records/phhq3wr65cap535",
			Body: strings.NewReader(`{
				"email":"new@example.com",
				"password":"12345678",
				"passwordConfirm":"12345678",
				"name":"test_name",
				"emailVisibility":true,
				"verified":true
			}`),
			RequestHeaders: map[string]string{
				// users, test@example.com
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"email":"new@example.com"`,
				`"name":"test_name"`,
				`"emailVisibility":true`,
				`"verified":true`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				record, _ := app.Dao().FindRecordById("nologin", "phhq3wr65cap535")
				if !record.ValidatePassword("12345678") {
					t.Fatal("Password update failed.")
				}
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "update auth record with valid data by admin",
			Method: http.MethodPatch,
			Url:    "/api/collections/users/records/2108356222582259712",
			Body: strings.NewReader(`{
				"username":"test.valid",
				"email":"new@example.com",
				"password":"12345678",
				"passwordConfirm":"12345678",
				"rel":"achvryl401bhse3",
				"emailVisibility":true,
				"verified":false
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"username":"test.valid"`,
				`"email":"new@example.com"`,
				`"rel":"achvryl401bhse3"`,
				`"emailVisibility":true`,
				`"verified":false`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				record, _ := app.Dao().FindRecordById("users", "2108356222582259712")
				if !record.ValidatePassword("12345678") {
					t.Fatal("Password update failed.")
				}
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "update auth record with valid data by guest (empty update filter)",
			Method: http.MethodPatch,
			Url:    "/api/collections/nologin/records/3480271880273794066",
			Body: strings.NewReader(`{
				"username":"test_new",
				"emailVisibility":true,
				"name":"test"
			}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"username":"test_new"`,
				`"email":"test@example.com"`, // the email should be visible since we updated the emailVisibility
				`"emailVisibility":true`,
				`"verified":false`,
				`"name":"test"`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "success password change with oldPassword",
			Method: http.MethodPatch,
			Url:    "/api/collections/nologin/records/3480271880273794066",
			Body: strings.NewReader(`{
				"password":"123456789",
				"passwordConfirm":"123456789",
				"oldPassword":"1234567890"
			}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"3480271880273794066"`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"password"`,
				`"passwordConfirm"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelAfterUpdate":          1,
				"OnModelBeforeUpdate":         1,
				"OnRecordAfterUpdateRequest":  1,
				"OnRecordBeforeUpdateRequest": 1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				record, _ := app.Dao().FindRecordById("nologin", "3480271880273794066")
				if !record.ValidatePassword("123456789") {
					t.Fatal("Password update failed.")
				}
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type RecordCrudTestSuite struct {
	suite.Suite
	App            *tests.TestApp
	AdminAuthToken string
	UserAuthToken  string
}

func (suite *RecordCrudTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.AdminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAyMzYxMTQsImlkIjoiMjEwNzk3NzEyNzUyODc1OTI5NiIsInR5cGUiOiJhZG1pbiJ9.ikCEJR-iPIrZwpbsWjtslMdq75suCAEYfaRK7Oz-NZ0"
	suite.UserAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiIyMTA3OTc3Mzk3MDYzMTIyOTQ0IiwiZXhwIjoxNzMwOTEyMTQzLCJpZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwidmVyaWZpZWQiOnRydWV9.Us_731ziRkeeZvYvXiXsc6CKEwdKp4rSvsGbG5L1OUQ"
	suite.App = app
}

func (suite *RecordCrudTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestRecordCrudTestSuite(t *testing.T) {
	suite.Run(t, new(RecordCrudTestSuite))
}
