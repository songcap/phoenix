/********************************************
我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。
为什么，应该怎么做请写出代码？

答：我认为还是有必要wrap这个error的，理由如下
1. 从信息获取角度 在实际的业务当中，经常会出现一次请求服务端需要处理两张表。如果不加上错误的上下文信息，上层就无法判断是哪个库哪个表没有接收到
2. 从简化代码逻辑的角度 如果在dao层出现问题打日志，上层也需要打日志，因为有可能多个dao层的接口被不同的上层调用，所以还是没办法作区分。
3. 从代码复用的角度来说 可以服用之前定义好的错误，因为wrap也会保留原始的error

以下例子以mongodb为例
*********************************************/
package main

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

const (
	database        = "p_testdb"
	collection_user = "p_userschema"
	collection_task = "p_taskschema"
)

type UserSchema struct {
	UserId string `bson:"userid"`
	Phone  string `bson:"phone"`
	Job    int    `bson:"job"`
}

type TaskSchema struct {
	Category int      `bson:"category"`
	Proj     string   `bson:"proj"`
	UserIds  []string `bson:"userids"`
}

/***
两个表出现NotFound都会导致流程失败的场景
给user派活
需要user和task都有记录
***/
func Handle2Schema(phone, proj string) error {
	user := &UserSchema{}
	err := FindOne(database, collection_user, bson.M{"phone": phone}, user)
	if err != nil {
		return errors.Wrap(err, "[database:"+database+" coll:"+collection_user+" find phone: "+phone+" failed]")
	}

	task := &TaskSchema{}
	err = FindOne(database, collection_task, bson.M{"proj": proj}, task)
	if err != nil {
		return errors.Wrap(err, "[database:"+database+" coll:"+collection_task+" find proj: "+proj+" failed]")
	}

	//insert
	err = InsertOne(database, collection_task, bson.M{"$push": bson.M{"userids": user.UserId}})
	if err != nil {
		return errors.Wrap(err, "[database:"+database+" coll:"+collection_task+" insert user: "+user.UserId+" failed]")
	}
	return nil
}

//------------------------------------------------------------------
//dao上层的业务接口代码段
func _AssignTask2UserUnwrap(phone, proj string) error {
	//第一种想知道具体的error是什么，并且用哨兵的方式做判断和处理
	if err := Handle2Schema(phone, proj); err != nil {
		//注：其实这里面Unwrap是不能直接用的，是给标准库的errors实现的.. 用Cause才能达到预期的结果
		if errors.Unwrap(err) == mgo.ErrNotFound {
			return errors.Wrap(err, "项目或用户不存在")
		} else if mgo.IsDup(errors.Unwrap(err)) {
			return errors.Wrap(err, "该用户已在项目组")
		} else {
			return errors.Wrap(err, "数据库异常")
		}
	}

	return nil
}

func _AssignTask2UserAsIs(phone, proj string) error {
	//使用As或者Is的方式判断error

	if err := Handle2Schema(phone, proj); err != nil {
		//使用As或者Is判断一下，err是否在里面
		if errors.Is(err, mgo.ErrNotFound) {
			//用errors.WithMessage添加信息,
			return errors.WithMessage(err, "项目或用户不存在")
		} else if mgo.IsDup(errors.Cause(err)) {
			return errors.WithMessagef(err, "该用户%s已在项目组", phone)
		} else {
			return errors.WithMessage(err, "数据库异常")
		}
	}

	return nil
}

// output like this ----------------------- _AssignTask2UserAsIs

// main.Handle2Schema
//         ../phoenix/error_handling/homework/pkg_errors.go:54
// main._AssignTask2UserAsIs
//         ../phoenix/error_handling/homework/pkg_errors.go:84
// main.main
//         ../phoenix/error_handling/homework/pkg_errors.go:102
// runtime.main
//         /usr/lib/golang/src/runtime/proc.go:204
// runtime.goexit
//         /usr/lib/golang/src/runtime/asm_amd64.s:1374
// 项目或用户不存在
// main.main
//         ../phoenix/error_handling/homework/pkg_errors.go:104
// runtime.main
//         /usr/lib/golang/src/runtime/proc.go:204
// runtime.goexit
//         /usr/lib/golang/src/runtime/asm_amd64.s:1374

// output like this ----------------------- _AssignTask2UserUnwrap
// not found
// [database:p_testdb coll:p_taskschema find proj: 10 failed]
// main.Handle2Schema
//         ../phoenix/error_handling/homework/pkg_errors.go:54
// main._AssignTask2UserUnwrap
//         ../phoenix/error_handling/homework/pkg_errors.go:68
// main.main
//         ../phoenix/error_handling/homework/pkg_errors.go:101
// runtime.main
//         /usr/lib/golang/src/runtime/proc.go:204
// runtime.goexit
//         /usr/lib/golang/src/runtime/asm_amd64.s:1374
// 数据库异常
// main._AssignTask2UserUnwrap
//         ../phoenix/error_handling/homework/pkg_errors.go:74
// main.main
//         ../phoenix/error_handling/homework/pkg_errors.go:101
// runtime.main
//         /usr/lib/golang/src/runtime/proc.go:204
// runtime.goexit
//         /usr/lib/golang/src/runtime/asm_amd64.s:1374
// main.main
//         ../phoenix/error_handling/homework/pkg_errors.go:104
// runtime.main
//         /usr/lib/golang/src/runtime/proc.go:204
// runtime.goexit
//         /usr/lib/golang/src/runtime/asm_amd64.s:1374
