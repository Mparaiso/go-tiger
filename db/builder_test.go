package db_test

import (
	"testing"

	"github.com/Mparaiso/go-tiger/db"
	. "github.com/Mparaiso/go-tiger/db/expression"
	"github.com/Mparaiso/go-tiger/db/platform"
	"github.com/Mparaiso/go-tiger/test"
)

type TestDatabasePlatform struct{}

func (platform *TestDatabasePlatform) ModifyLimitQuery(query string, maxResult int, firstResult int) string {
	return ""
}

func (platform *TestDatabasePlatform) SetParent(platform.DatabasePlatform)  {}
func (platform *TestDatabasePlatform) GetParent() platform.DatabasePlatform { return nil }

type TestConnection struct {
	platform.DatabasePlatform
	db.Connection
}

func (c *TestConnection) GetDatabasePlatform() platform.DatabasePlatform {
	return c.DatabasePlatform
}

func NewTestConnection(t *testing.T) db.Connection {
	return &TestConnection{DatabasePlatform: &TestDatabasePlatform{}}
}

type Connection interface{}

func TestBuilderSelect(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.id").
		From("users", "u")
	test.Fatal(t, qb.String(), "SELECT u.id FROM users u")
}

func TestBuilderSelectWithWhere(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.id").
		From("users", "u").
		Where(And(Eq("u.nickname", "?")))
	test.Fatal(t, qb.String(), "SELECT u.id FROM users u WHERE u.nickname = ?")
}

func TestBuilderSelectWithLeftJoin(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		LeftJoin("u", "phones", "p", Eq("p.user_id", "u.id"))
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u LEFT JOIN phones p ON p.user_id = u.id")

}

func TestBuilderSelectWithJoin(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").From("users", "u").
		Join("u", "phones", "p", Eq("p.user_id", "u.id"))
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u JOIN phones p ON p.user_id = u.id")

}

func TestBuilderSelectWithInnerJoin(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		InnerJoin("u", "phones", "p", Eq("p.user_id", "u.id"))
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u INNER JOIN phones p ON p.user_id = u.id")

}

func TestBuilderSelectWithRightJoin(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		RightJoin("u", "phones", "p", Eq("p.user_id", "u.id"))
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u RIGHT JOIN phones p ON p.user_id = u.id")

}

func TestBuilderSelectWithAndWhereConditions(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		Where("u.username = ?").
		AndWhere("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u WHERE (u.username = ?) AND (u.name = ?)")
}

func TestSelectWithOrWhereConditions(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		Where("u.username = ?").
		OrWhere("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u WHERE (u.username = ?) OR (u.name = ?)")
}

func TestSelectWithOrOrWhereConditions(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		OrWhere("u.username = ?").
		OrWhere("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u WHERE (u.username = ?) OR (u.name = ?)")
}

func TestSelectWithOrOrOrWhereConditions(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		OrWhere("u.username = ?").
		OrWhere("u.name = ?").
		OrWhere("u.age = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u WHERE (u.username = ?) OR (u.name = ?) OR (u.age = ?)")
}

func TestSelectWithAndOrWhereConditions(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		Where("u.username = ?").
		AndWhere("u.username = ?").
		OrWhere("u.name = ?").
		AndWhere("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u WHERE (((u.username = ?) AND (u.username = ?)) OR (u.name = ?)) AND (u.name = ?)")
}

func TestSelectGroupBy(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id")
}

func TestSelectEmptyGroupBy(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)

	qb.Select("u.*", "p.*").
		GroupBy().
		From("users", "u")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u")
}
func TestSelectEmptyAddGroupBy(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		AddGroupBy().
		From("users", "u")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u")
}
func TestSelectAddGroupBy(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		AddGroupBy("u.foo")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id, u.foo")
}

func TestSelectAddGroupBys(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		AddGroupBy("u.foo", "u.bar")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id, u.foo, u.bar")
}

func TestSelectHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)

	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		Having("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING u.name = ?")
}
func TestSelectAndHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		AndHaving("u.name = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING u.name = ?")
}
func TestSelectHavingAndHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		Having("u.name = ?").
		AndHaving("u.username = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING (u.name = ?) AND (u.username = ?)")
}
func TestSelectHavingOrHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		Having("u.name = ?").
		OrHaving("u.username = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING (u.name = ?) OR (u.username = ?)")
}
func TestSelectOrHavingOrHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		OrHaving("u.name = ?").
		OrHaving("u.username = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING (u.name = ?) OR (u.username = ?)")
}
func TestSelectHavingAndOrHaving(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		GroupBy("u.id").
		Having("u.name = ?").
		OrHaving("u.username = ?").
		AndHaving("u.username = ?")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u GROUP BY u.id HAVING ((u.name = ?) OR (u.username = ?)) AND (u.username = ?)")
}
func TestSelectOrderBy(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		OrderBy("u.name")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u ORDER BY u.name ASC")
}
func TestSelectAddOrderBy(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*", "p.*").
		From("users", "u").
		OrderBy("u.name").
		AddOrderBy("u.username", "DESC")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u ORDER BY u.name ASC, u.username DESC")
}
func TestSelectAddAddOrderBy(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)

	qb.Select("u.*", "p.*").
		From("users", "u").
		AddOrderBy("u.name").
		AddOrderBy("u.username", "DESC")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u ORDER BY u.name ASC, u.username DESC")
}
func TestEmptySelect(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	test.Fatal(t, qb.GetType(), db.Select)
}
func TestSelectAddSelect(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)

	qb.Select("u.*").
		AddSelect("p.*").
		From("users", "u")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u")
}
func TestEmptyAddSelect(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	test.Fatal(t, qb.GetType(), db.Select)
}
func TestSelectMultipleFrom(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.*").
		AddSelect("p.*").
		From("users", "u").
		From("phonenumbers", "p")
	test.Fatal(t, qb.String(), "SELECT u.*, p.* FROM users u, phonenumbers p")
}

func TestUpdate(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Update("users", "u").
		Set("u.foo", "?").
		Set("u.bar", "?")
	test.Fatal(t, qb.GetType(), db.Update)
	test.Fatal(t, qb.String(), "UPDATE users u SET u.foo = ?, u.bar = ?")
}

func TestUpdateWithoutAlias(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Update("users").
		Set("foo", "?").
		Set("bar", "?")
	test.Fatal(t, qb.String(), "UPDATE users SET foo = ?, bar = ?")
}

func TestUpdateWhere(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Update("users", "u").
		Set("u.foo", "?").
		Where("u.foo = ?")
	test.Fatal(t, qb.String(), "UPDATE users u SET u.foo = ? WHERE u.foo = ?")
}

func TestEmptyUpdate(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection).Update("")
	test.Fatal(t, qb.GetType(), db.Update)
}

func TestDelete(t *testing.T) {

	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Delete("users", "u")
	test.Fatal(t, db.Delete, qb.GetType())
	test.Fatal(t, qb.String(), "DELETE FROM users u")
}
func TestDeleteWithoutAlias(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Delete("users")
	test.Fatal(t, db.Delete, qb.GetType())
	test.Fatal(t, "DELETE FROM users", qb.String())
}
func TestDeleteWhere(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Delete("users", "u").
		Where("u.foo = ?")
	test.Fatal(t, "DELETE FROM users u WHERE u.foo = ?", qb.String())
}
func TestInsertValues(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Insert("users").
		SetValue("foo", "?").
		SetValue("bar", "?")
	test.Fatal(t, db.Insert, qb.GetType())
	test.Fatal(t, "INSERT INTO users (foo, bar) VALUES(?, ?)", qb.String())
}
func TestInsertReplaceValues(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Insert("users").
		SetValue("bar", "?").
		SetValue("foo", "?")
	test.Fatal(t, db.Insert, qb.GetType())
	test.Fatal(t, "INSERT INTO users (bar, foo) VALUES(?, ?)", qb.String())
}
func TestInsertSetValue(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Insert("users").
		SetValue("foo", "bar").
		SetValue("bar", "?").
		SetValue("foo", "?")
	test.Fatal(t, db.Insert, qb.GetType())
	test.Fatal(t, "INSERT INTO users (foo, bar) VALUES(?, ?)", qb.String())
}
func TestInsertValuesSetValue(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	qb.Insert("users").
		SetValue("foo", "?").
		SetValue("bar", "?")
	test.Fatal(t, db.Insert, qb.GetType())
	test.Fatal(t, "INSERT INTO users (foo, bar) VALUES(?, ?)", qb.String())
}

func TestGetConnection(t *testing.T) {
	connection := NewTestConnection(t)

	qb := db.NewQueryBuilder(connection)
	test.Fatal(t, connection, qb.GetConnection())
}

func TestGetState(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	test.Fatal(t, qb.GetState(), db.Clean)
	qb.Select("u.*").From("users", "u")
	test.Fatal(t, qb.GetState(), db.Dirty)
	qb.String()
	test.Fatal(t, qb.GetState(), db.Clean)
}

func TestSetMaxResults(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.SetMaxResults(10)
	test.Fatal(t, qb.GetState(), db.Dirty)
	test.Fatal(t, qb.GetMaxResults(), 10)
}
func TestSetFirstResult(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.SetFirstResult(10)
	test.Fatal(t, qb.GetState(), db.Dirty)
	test.Fatal(t, qb.GetFirstResult(), 10)
}

/*
 func TestResetQueryPart (t *testing.T) {
{
    qb   = db.NewQueryBuilder(connection)
    qb.Select("u.*").From("users", "u").where("u.name = ?")
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u WHERE u.name = ?")
    qb.resetQueryPart("where")
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u")
}
 func TestResetQueryParts (t *testing.T) {
{
    qb   = db.NewQueryBuilder(connection)
    qb.Select("u.*").From("users", "u").where("u.name = ?").orderBy("u.name")
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u WHERE u.name = ? ORDER BY u.name ASC")
    qb.resetQueryParts(array("where", "orderBy"))
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u")
}
 func TestCreateNamedParameter (t *testing.T) {
{
    qb   = db.NewQueryBuilder(connection)
    qb.Select("u.*").From("users", "u").where(
        qb.expr().eq("u.name", qb.createNamedParameter(10, \PDO::PARAM_INT))
    )
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u WHERE u.name = :dcValue1")
    test.Fatal(t,qb.String(),10, qb.getParameter("dcValue1"))
    test.Fatal(t,qb.String(),\PDO::PARAM_INT, qb.getParameterType("dcValue1"))
}
 func TestCreateNamedParameterCustomPlaceholder (t *testing.T) {
{
    qb   = db.NewQueryBuilder(connection)
    qb.Select("u.*").From("users", "u").where(
        qb.expr().eq("u.name", qb.createNamedParameter(10, \PDO::PARAM_INT, ":test"))
    )
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u WHERE u.name = :test")
    test.Fatal(t,qb.String(),10, qb.getParameter("test"))
    test.Fatal(t,qb.String(),\PDO::PARAM_INT, qb.getParameterType("test"))
}
 func TestCreatePositionalParameter (t *testing.T) {
{
    qb   = db.NewQueryBuilder(connection)
    qb.Select("u.*").From("users", "u").where(
        qb.expr().eq("u.name", qb.createPositionalParameter(10, \PDO::PARAM_INT))
    )
    test.Fatal(t,qb.String(),"SELECT u.* FROM users u WHERE u.name = ?")
    test.Fatal(t,qb.String(),10, qb.getParameter(1))
    test.Fatal(t,qb.String(),\PDO::PARAM_INT, qb.getParameterType(1))
}
/**
 * @group DBAL-172
*/
//      func TestReferenceJoinFromJoin (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("COUNT(DISTINCT news.id)")
//             .From("cb_newspages", "news")
//             .InnerJoin("news", "nodeversion", "nv", "nv.refId = news.id AND nv.refEntityname=\"News\'')
//             .InnerJoin("invalid", "nodetranslation", "nt", "nv.nodetranslation = nt.id")
//             .InnerJoin("nt", "node", "n", "nt.node = n.id")
//             .where("nt.lang = :lang AND n.deleted != 1")
//         this.setExpectedException("Doctrine\DBAL\Query\QueryException", "The given alias "invalid" is not part of any FROM or JOIN clause table. The currently registered aliases are: news, nv.")
//         test.Fatal(t,qb.String(),'', qb.getSQL())
//     }
//     /**
//      * @group DBAL-172
//      */
//      func TestSelectFromMasterWithWhereOnJoinedTables (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("COUNT(DISTINCT news.id)")
//             .From("newspages", "news")
//             .InnerJoin("news", "nodeversion", "nv", "nv.refId = news.id AND nv.refEntityname="Entity\\News"")
//             .InnerJoin("nv", "nodetranslation", "nt", "nv.nodetranslation = nt.id")
//             .InnerJoin("nt", "node", "n", "nt.node = n.id")
//             .where("nt.lang = ?")
//             .andWhere("n.deleted = 0")
//         test.Fatal(t,qb.String(),"SELECT COUNT(DISTINCT news.id) FROM newspages news INNER JOIN nodeversion nv ON nv.refId = news.id AND nv.refEntityname="Entity\\News" INNER JOIN nodetranslation nt ON nv.nodetranslation = nt.id INNER JOIN node n ON nt.node = n.id WHERE (nt.lang = ?) AND (n.deleted = 0)", qb.getSQL())
//     }
//     /**
//      * @group DBAL-442
//      */
//      func TestSelectWithMultipleFromAndJoins (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("DISTINCT u.id")
//             .From("users", "u")
//             .From("articles", "a")
//             .InnerJoin("u", "permissions", "p", "p.user_id = u.id")
//             .InnerJoin("a", "comments", "c", "c.article_id = a.id")
//             .where("u.id = a.user_id")
//             .andWhere("p.read = 1")
//         test.Fatal(t,qb.String(),"SELECT DISTINCT u.id FROM users u INNER JOIN permissions p ON p.user_id = u.id, articles a INNER JOIN comments c ON c.article_id = a.id WHERE (u.id = a.user_id) AND (p.read = 1)", qb.getSQL())
//     }
//     /**
//      * @group DBAL-774
//      */
func TestSelectWithJoinsWithMultipleOnConditionsParseOrder(t *testing.T) {
	t.Skip("join alias should also match alias in other joins if not found in from")
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("a.id").
		From("table_a", "a").
		Join("a", "table_b", "b", "a.fk_b = b.id").
		Join("b", "table_c", "c", "c.fk_b = b.id AND b.language = ?").
		Join("a", "table_d", "d", "a.fk_d = d.id").
		Join("c", "table_e", "e", "e.fk_c = c.id AND e.fk_d = d.id")
	test.FatalWithDiff(t, qb.String(),
		"SELECT a.id "+
			"FROM table_a a "+
			"INNER JOIN table_b b ON a.fk_b = b.id "+
			"INNER JOIN table_d d ON a.fk_d = d.id "+
			"INNER JOIN table_c c ON c.fk_b = b.id AND b.language = ? "+
			"INNER JOIN table_e e ON e.fk_c = c.id AND e.fk_d = d.id")
}

//     /**
//      * @group DBAL-774
//      */
//      func TestSelectWithMultipleFromsAndJoinsWithMultipleOnConditionsParseOrder (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("a.id")
//             .From("table_a", "a")
//             .From("table_f", "f")
//             .join("a", "table_b", "b", "a.fk_b = b.id")
//             .join("b", "table_c", "c", "c.fk_b = b.id AND b.language = ?")
//             .join("a", "table_d", "d", "a.fk_d = d.id")
//             .join("c", "table_e", "e", "e.fk_c = c.id AND e.fk_d = d.id")
//             .join("f", "table_g", "g", "f.fk_g = g.id")
//         test.Fatal(t,qb.String(),
//             "SELECT a.id " .
//             "FROM table_a a " .
//             "INNER JOIN table_b b ON a.fk_b = b.id " .
//             "INNER JOIN table_d d ON a.fk_d = d.id " .
//             "INNER JOIN table_c c ON c.fk_b = b.id AND b.language = ? " .
//             "INNER JOIN table_e e ON e.fk_c = c.id AND e.fk_d = d.id, " .
//             "table_f f " .
//             "INNER JOIN table_g g ON f.fk_g = g.id",
//             (string) qb
//         )
//     }
//      func TestClone (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("u.id")
//             .From("users", "u")
//             .where("u.id = :test")
//         qb.setParameter(":test", (object) 1)
//         qb_clone = clone qb
//         test.Fatal(t,qb.String(),(string) qb_clone)
//         qb.andWhere("u.id = 1")
//         this.assertFalse(qb.getQueryParts() === qb_clone.getQueryParts())
//         this.assertFalse(qb.getParameters() === qb_clone.getParameters())
//     }
func TestSimpleSelectWithoutTableAlias(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("id").
		From("users")
	test.Fatal(t, qb.String(), "SELECT id FROM users")
}

//      func TestSelectWithSimpleWhereWithoutTableAlias (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("id", "name")
//             .From("users")
//             .where("awesome=9001")
//         test.Fatal(t,qb.String(),"SELECT id, name FROM users WHERE awesome=9001")
//     }
func TestComplexSelectWithoutTableAliases(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("DISTINCT users.id").
		From("users").
		From("articles").
		InnerJoin("articles", "comments", "c", "c.article_id = articles.id").
		InnerJoin("users", "permissions", "p", "p.user_id = users.id").
		Where("users.id = articles.user_id").
		AndWhere("p.read = 1")
	test.FatalWithDiff(t, qb.String(),
		"SELECT DISTINCT users.id FROM users INNER JOIN permissions p ON p.user_id"+
			" = users.id, articles INNER JOIN comments c ON c.article_id = articles.id "+
			"WHERE (users.id = articles.user_id) AND (p.read = 1)")
}
func TestComplexSelectWithSomeTableAliases(t *testing.T) {
	// t.Skip("Skipped until the order of joins satisfies the doctrine specification")

	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("u.id").
		From("users", "u").
		From("articles").
		InnerJoin("u", "permissions", "p", "p.user_id = u.id").
		InnerJoin("articles", "comments", "c", "c.article_id = articles.id")
	test.Fatal(t, qb.String(),
		"SELECT u.id FROM users u INNER JOIN permissions p ON p.user_id = u.id, articles INNER JOIN comments c ON c.article_id = articles.id")
}
func TestSelectAllFromTableWithoutTableAlias(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("users.*").
		From("users")
	test.Fatal(t, qb.String(), "SELECT users.* FROM users")
}
func TestSelectAllWithoutTableAlias(t *testing.T) {
	connection := NewTestConnection(t)
	qb := db.NewQueryBuilder(connection)
	qb.Select("*").
		From("users")
	test.Fatal(t, qb.String(), "SELECT * FROM users")
}

//     /**
//      * @group DBAL-959
//      */
//      func TestGetParameterType (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("*").From("users")
//         this.assertNull(qb.getParameterType("name"))
//         qb.where("name = :name")
//         qb.setParameter("name", "foo")
//         this.assertNull(qb.getParameterType("name"))
//         qb.setParameter("name", "foo", \PDO::PARAM_STR)
//         this.assertSame(\PDO::PARAM_STR, qb.getParameterType("name"))
//     }
//     /**
//      * @group DBAL-959
//      */
//      func TestGetParameterTypes (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("*").From("users")
//         this.assertSame(array(), qb.getParameterTypes())
//         qb.where("name = :name")
//         qb.setParameter("name", "foo")
//         this.assertSame(array(), qb.getParameterTypes())
//         qb.setParameter("name", "foo", \PDO::PARAM_STR)
//         qb.where("is_active = :isActive")
//         qb.setParameter("isActive", true, \PDO::PARAM_BOOL)
//         this.assertSame(array("name" => \PDO::PARAM_STR, "isActive" => \PDO::PARAM_BOOL), qb.getParameterTypes())
//     }
//     /**
//      * @group DBAL-1137
//      */
//      func TestJoinWithNonUniqueAliasThrowsException (t *testing.T) {

//         qb := db.NewQueryBuilder(connection)
//         qb.Select("a.id")
//             .From("table_a", "a")
//             .join("a", "table_b", "a", "a.fk_b = a.id")
//         this.setExpectedException(
//             "Doctrine\DBAL\Query\QueryException",
//             "The given alias "a" is not unique in FROM and JOIN clause table. The currently registered aliases are: a."
//         )
//         qb.getSQL()
//     }
// }
// */
