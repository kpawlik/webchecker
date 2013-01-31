package checker

import (
	"appengine"
	"appengine/datastore"
)

type Constuctor func() Keyed

type Keyed interface {
	key(appengine.Context, *datastore.Key) *datastore.Key
	newKey(appengine.Context, *datastore.Key) *datastore.Key
	tableName() string
}

type DB struct {
	ctx appengine.Context
}

func NewDB(c appengine.Context) *DB {
	return &DB{c}
}

func (db *DB) LastRec(constr Constuctor, fieldName string, parent *datastore.Key) (Keyed, error) {
	ex := constr()
	q := datastore.NewQuery(ex.tableName()).
		Limit(1).
		Order(fieldName).Ancestor(parent)
	for t := q.Run(db.ctx); ; {
		_, err := t.Next(ex)
		if err == datastore.Done {
			return ex, nil
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (db *DB) Get(obj Keyed, parent *datastore.Key) (res Keyed) {
	var (
		err error
	)
	key := obj.key(db.ctx, parent)
	if err = datastore.Get(db.ctx, key, obj); err == datastore.ErrNoSuchEntity {
		return
	}
	if err != nil {
		panic(err)
	}
	res = obj
	return
}

func (db *DB) SaveNew(obj Keyed, parent *datastore.Key) (res Keyed, err error) {
	if res = db.Get(obj, parent); res != nil {
		p("GET - ", res)
		return
	}
	_, err = datastore.Put(db.ctx, obj.key(db.ctx, parent), obj)
	return obj, err
}

func (db *DB) Save(obj Keyed, parent *datastore.Key) (res Keyed, err error) {
	_, err = datastore.Put(db.ctx, obj.key(db.ctx, parent), obj)
	return obj, err
}

//////////////////////////
func nCheckResult() Keyed {
	return &CheckResult{}
}

func (cr *CheckResult) newKey(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, cr.tableName(), parent)
}

func (cr *CheckResult) key(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, cr.tableName(), "", 0, parent)
}

func (cr *CheckResult) tableName() string {
	return "CheckResult"
}

/////////////////////

func nUser() Keyed {
	return &User{}
}

func (cr *User) newKey(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, cr.tableName(), parent)
}

func (cr *User) key(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, cr.tableName(), cr.Name, 0, parent)
}

func (cr *User) tableName() string {
	return "Usr"
}

/////////////////////

func nConfig() Keyed {
	return &Config{}
}

func (cr *Config) newKey(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, cr.tableName(), parent)
}

func (cr *Config) key(c appengine.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, cr.tableName(), "", 0, parent)
}

func (cr *Config) tableName() string {
	return "Cfg"
}
