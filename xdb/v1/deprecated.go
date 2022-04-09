//go:generate msgp -tests=false
//msgp:ignore Repository
package v1

// Deprecated
// 请使用 xdb.Model
type Entity = Model

// Deprecated
// 请使用 xdb.Dao
type Repository = Dao
