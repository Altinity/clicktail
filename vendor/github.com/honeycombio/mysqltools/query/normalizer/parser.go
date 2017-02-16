package normalizer

import (
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/sqlparser"
)

type Parser struct {
	LastStatement string
	LastTables    []string
	LastComments  []string
}

func (n *Parser) NormalizeQuery(q string) string {
	n.LastStatement = ""
	n.LastTables = make([]string, 0)
	n.LastComments = make([]string, 0)

	if q == "" {
		return ""
	}

	q = strings.ToLower(q)

	sqlAST, err := sqlparser.Parse(q)
	if err != nil {
		logrus.WithError(err).Debug("parse error, falling back to scan, query: ", q)
		s := &Scanner{}
		return s.NormalizeQuery(q)
	}

	newAST := transform(sqlAST, n)
	if newAST == nil {
		return ""
	}

	n.LastStatement = classifyStatement(sqlAST)

	var lastTables []string
	for _, t := range n.LastTables {
		lastTables = append(lastTables, strings.Trim(t, "`"))
	}

	sort.Sort(sort.StringSlice(lastTables))

	n.LastTables = lastTables

	return string(sqlparser.Serialize(newAST, len(q)))
}

// QuestionMarkExpr is a special SQLNode used to render '?'.  we replace literal values with this in our transformer
type QuestionMarkExpr struct {
}

func (q *QuestionMarkExpr) Format(buf *sqlparser.TrackedBuffer) {
	buf.Myprintf("?")
}

func (q *QuestionMarkExpr) Serialize(runes []rune) []rune {
	return append(runes, '?')
}

func (*QuestionMarkExpr) IExpr()    {}
func (*QuestionMarkExpr) IValExpr() {}

func (n *Parser) TransformSelect(node *sqlparser.Select) sqlparser.SQLNode {
	n.addComments(node.Comments)
	node.Comments = removeComments(node.Comments)
	node.SelectExprs, _ = transform(node.SelectExprs, n).(sqlparser.SelectExprs)
	node.Where, _ = transform(node.Where, n).(*sqlparser.Where)
	node.From, _ = transform(node.From, n).(sqlparser.TableExprs)
	node.Limit, _ = transform(node.Limit, n).(*sqlparser.Limit)
	return node
}
func (n *Parser) TransformSelectExprs(node sqlparser.SelectExprs) sqlparser.SQLNode {
	var newSlice sqlparser.SelectExprs
	for _, se := range node {
		selectExpr, _ := transform(se, n).(sqlparser.SelectExpr)
		newSlice = append(newSlice, selectExpr)
	}
	return newSlice
}
func (n *Parser) TransformUnion(node *sqlparser.Union) sqlparser.SQLNode {
	node.Left, _ = transform(node.Left, n).(sqlparser.SelectStatement)
	node.Right, _ = transform(node.Right, n).(sqlparser.SelectStatement)
	return node
}
func (n *Parser) TransformInsert(node *sqlparser.Insert) sqlparser.SQLNode {
	n.addComments(node.Comments)
	node.Comments = removeComments(node.Comments)
	node.Table, _ = transform(node.Table, n).(*sqlparser.TableName)
	node.Rows, _ = transform(node.Rows, n).(sqlparser.InsertRows)
	node.OnDup, _ = transform(sqlparser.UpdateExprs(node.OnDup), n).(sqlparser.OnDup)
	return node
}
func (n *Parser) TransformUpdate(node *sqlparser.Update) sqlparser.SQLNode {
	n.addComments(node.Comments)
	node.Comments = removeComments(node.Comments)
	node.Table, _ = transform(node.Table, n).(*sqlparser.TableName)
	node.Exprs, _ = transform(node.Exprs, n).(sqlparser.UpdateExprs)
	node.Where, _ = transform(node.Where, n).(*sqlparser.Where)
	node.Limit, _ = transform(node.Limit, n).(*sqlparser.Limit)
	return node
}
func (n *Parser) TransformUpdateExprs(node sqlparser.UpdateExprs) sqlparser.SQLNode {
	var newSlice sqlparser.UpdateExprs
	for _, ue := range node {
		updateExpr, _ := transform(ue, n).(*sqlparser.UpdateExpr)
		newSlice = append(newSlice, updateExpr)
	}
	return newSlice
}
func (n *Parser) TransformDelete(node *sqlparser.Delete) sqlparser.SQLNode {
	n.addComments(node.Comments)
	node.Comments = removeComments(node.Comments)
	node.Table, _ = transform(node.Table, n).(*sqlparser.TableName)
	node.Where, _ = transform(node.Where, n).(*sqlparser.Where)
	node.Limit, _ = transform(node.Limit, n).(*sqlparser.Limit)
	return node
}
func (n *Parser) TransformSet(node *sqlparser.Set) sqlparser.SQLNode {
	n.addComments(node.Comments)
	node.Comments = removeComments(node.Comments)
	node.Exprs, _ = transform(node.Exprs, n).(sqlparser.UpdateExprs)
	return node
}

func (n *Parser) TransformDDL(node *sqlparser.DDL) sqlparser.SQLNode {
	return node
}

func (n *Parser) TransformColumnDefinition(node *sqlparser.ColumnDefinition) sqlparser.SQLNode {
	return node
}

func (n *Parser) TransformCreateTable(node *sqlparser.CreateTable) sqlparser.SQLNode {
	n.addTableName(string(node.Name))
	return node
}

func (n *Parser) TransformStarExpr(node *sqlparser.StarExpr) sqlparser.SQLNode {
	if len(node.TableName) > 0 {
		n.addTableName(string(node.TableName))
	}
	return node
}

func (n *Parser) TransformNonStarExpr(node *sqlparser.NonStarExpr) sqlparser.SQLNode {
	node.Expr, _ = transform(node.Expr, n).(sqlparser.Expr)
	return node
}

func (n *Parser) TransformAliasedTableExpr(node *sqlparser.AliasedTableExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.SimpleTableExpr)
	return node
}

func (n *Parser) TransformTableName(node *sqlparser.TableName) sqlparser.SQLNode {
	if node == nil {
		return nil
	}

	n.addTableName(sqlparser.String(node))
	return node
}

func (n *Parser) TransformParenTableExpr(node *sqlparser.ParenTableExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.TableExpr)
	return node
}

func (n *Parser) TransformJoinTableExpr(node *sqlparser.JoinTableExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.LeftExpr, _ = transform(node.LeftExpr, n).(sqlparser.TableExpr)
	node.RightExpr, _ = transform(node.RightExpr, n).(sqlparser.TableExpr)
	node.On, _ = transform(node.On, n).(sqlparser.BoolExpr)
	return node
}
func (n *Parser) TransformIndexHints(node *sqlparser.IndexHints) sqlparser.SQLNode /* needed? */ {
	return node
}
func (n *Parser) TransformWhere(node *sqlparser.Where) sqlparser.SQLNode {
	if node == nil {
		return nil
	}

	node.Expr, _ = transform(node.Expr, n).(sqlparser.BoolExpr)
	return node
}
func (n *Parser) TransformAndExpr(node *sqlparser.AndExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Left, _ = transform(node.Left, n).(sqlparser.BoolExpr)
	node.Right, _ = transform(node.Right, n).(sqlparser.BoolExpr)
	return node
}
func (n *Parser) TransformOrExpr(node *sqlparser.OrExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Left, _ = transform(node.Left, n).(sqlparser.BoolExpr)
	node.Right, _ = transform(node.Right, n).(sqlparser.BoolExpr)
	return node
}
func (n *Parser) TransformNotExpr(node *sqlparser.NotExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.BoolExpr)
	return node
}
func (n *Parser) TransformParenBoolExpr(node *sqlparser.ParenBoolExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	newExpr, _ := transform(node.Expr, n).(sqlparser.BoolExpr)
	node.Expr = newExpr
	return node
}
func (n *Parser) TransformComparisonExpr(node *sqlparser.ComparisonExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Left, _ = transform(node.Left, n).(sqlparser.ValExpr)
	node.Right, _ = transform(node.Right, n).(sqlparser.ValExpr)
	return node
}
func (n *Parser) TransformRangeCond(node *sqlparser.RangeCond) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Left, _ = transform(node.Left, n).(sqlparser.ValExpr)
	node.From, _ = transform(node.From, n).(sqlparser.ValExpr)
	node.To, _ = transform(node.To, n).(sqlparser.ValExpr)
	return node
}
func (n *Parser) TransformExistsExpr(node *sqlparser.ExistsExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Subquery, _ = transform(node.Subquery, n).(*sqlparser.Subquery)
	return node
}
func (n *Parser) TransformTimestampVal(node sqlparser.TimestampVal) sqlparser.SQLNode {
	return &QuestionMarkExpr{}
}
func (n *Parser) TransformBinaryVal(node sqlparser.BinaryVal) sqlparser.SQLNode {
	return &QuestionMarkExpr{}
}
func (n *Parser) TransformStrVal(node sqlparser.StrVal) sqlparser.SQLNode {
	return &QuestionMarkExpr{}
}
func (n *Parser) TransformNumVal(node sqlparser.NumVal) sqlparser.SQLNode {
	return &QuestionMarkExpr{}
}
func (n *Parser) TransformValArg(node *sqlparser.ValArg) sqlparser.SQLNode {
	return &QuestionMarkExpr{}
}
func (n *Parser) TransformValTuple(node sqlparser.ValTuple) sqlparser.SQLNode {
	var newSlice sqlparser.ValTuple
	for _, val := range node {
		valExpr, _ := transform(val, n).(sqlparser.ValExpr)
		newSlice = append(newSlice, valExpr)
	}
	return newSlice
}
func (n *Parser) TransformNullVal(node *sqlparser.NullVal) sqlparser.SQLNode {
	return node
}
func (n *Parser) TransformColName(node *sqlparser.ColName) sqlparser.SQLNode {
	if node.Qualifier != nil {
		quals := strings.Split(string(node.Qualifier), ".")
		n.addTableName(quals[len(quals)-1])
	}

	return node
}
func (n *Parser) TransformSubquery(node *sqlparser.Subquery) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Select, _ = transform(node.Select, n).(sqlparser.SelectStatement)
	return node
}
func (n *Parser) TransformBinaryExpr(node *sqlparser.BinaryExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Left, _ = transform(node.Left, n).(sqlparser.Expr)
	node.Right, _ = transform(node.Right, n).(sqlparser.Expr)
	return node
}
func (n *Parser) TransformUnaryExpr(node *sqlparser.UnaryExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.Expr)
	return node
}
func (n *Parser) TransformFuncExpr(node *sqlparser.FuncExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Exprs, _ = transform(node.Exprs, n).(sqlparser.SelectExprs)
	return node
}
func (n *Parser) TransformCaseExpr(node *sqlparser.CaseExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.ValExpr)
	for i, _ := range node.Whens {
		node.Whens[i], _ = transform(node.Whens[i], n).(*sqlparser.When)
	}
	node.Else, _ = transform(node.Else, n).(sqlparser.ValExpr)
	return node
}
func (n *Parser) TransformWhen(node *sqlparser.When) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Cond, _ = transform(node.Cond, n).(sqlparser.BoolExpr)
	node.Val, _ = transform(node.Val, n).(sqlparser.ValExpr)
	return node
}
func (n *Parser) TransformOrder(node *sqlparser.Order) sqlparser.SQLNode {
	return node
}
func (n *Parser) TransformLimit(node *sqlparser.Limit) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Offset, _ = transform(node.Offset, n).(sqlparser.ValExpr)
	node.Rowcount, _ = transform(node.Rowcount, n).(sqlparser.ValExpr)
	return node
}
func (n *Parser) TransformUpdateExpr(node *sqlparser.UpdateExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.ValExpr)
	return node
}

func (n *Parser) TransformValues(node sqlparser.Values) sqlparser.SQLNode {
	var newSlice sqlparser.Values
	for _, rt := range node {
		rowTuple, _ := transform(rt, n).(sqlparser.RowTuple)
		newSlice = append(newSlice, rowTuple)
	}
	return newSlice
}

func (n *Parser) TransformTableExprs(node sqlparser.TableExprs) sqlparser.SQLNode {
	var newSlice sqlparser.TableExprs
	for _, te := range node {
		te, _ := transform(te, n).(sqlparser.TableExpr)
		newSlice = append(newSlice, te)
	}
	return newSlice
}

func (n *Parser) TransformAliasedTablExpr(node *sqlparser.AliasedTableExpr) sqlparser.SQLNode {
	if node == nil {
		return nil
	}
	node.Expr, _ = transform(node.Expr, n).(sqlparser.SimpleTableExpr)
	return node
}

func (n *Parser) addComments(comments sqlparser.Comments) {
	if comments == nil {
		return
	}
	for _, c := range comments {
		n.LastComments = append(n.LastComments, strings.TrimSpace(string(c)))
	}
}

func removeComments(comments sqlparser.Comments) sqlparser.Comments {
	if len(comments) == 0 {
		return comments
	}
	return make([][]rune, 0)
}

func (n *Parser) addTableName(tableNameStr string) {
	found := false
	for _, t := range n.LastTables {
		if t == tableNameStr {
			found = true
			break
		}
	}

	if !found {
		n.LastTables = append(n.LastTables, tableNameStr)
	}
}

func classifyStatement(node sqlparser.SQLNode) string {
	if node == nil {
		return ""
	}

	nodeType := reflect.TypeOf(node)
	switch nodeType {
	case selectType:
		return "select"
	case unionType:
		return "union"
	case insertType:
		return "insert"
	case updateType:
		return "update"
	case deleteType:
		return "delete"
	case setType:
		return "set"
	case createTableType:
		return "create table"
	case otherType:
		return ""
	case ddlType:
		return ""
	default:
		log.Printf("classifyStatement doesn't handle %+v", reflect.TypeOf(node))
		return ""
	}
}
