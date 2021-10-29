/***************************************
使用handle check的方式做错误处理 这种方案好像在2.0已经被否决了，因为也不怎么好用

我们可以将前面的代码进行化简，较少if err != nil的出现频率，
并统一在 handle代码块中对错误进行处理。
这种使用 check 和 handle 的方式会当 err 发生时，直接进入 check 关键字上方 最近的一个 handle err 块进行错误处理。在官方的这个例子中其实就已经发生了语言上模棱两可的地方， 当函数最下方的 w.Close 产生调用时， 上方与其最近的一个 handle err 还会再一次调用 w.Close，这其实是多余的。

此外，这种方式看似对代码进行了简化，但仔细一看这种方式与 defer 函数进行错误处理之间 减少了 if err != nil { return err  } 出现的频率，并没有带来任何本质区别。
****************************************/

package main 

import (
	"fmt"
)

func HandleAndCheck(phone, proj string) error {
	
	coll := ""

	handle err {
		//好处就是当有很多错误都是类似的就可以合并了
		return errors.Wrap(err, "[database:"+ database + " coll:"+ coll + " find "+phone+" or "+proj + "failed]")
	}
	
	user := &UserSchema{}
	coll = collection_user
	check FindOne(database, coll, bson.M{"phone": phone}, user)

	task := &TaskSchema{}
	coll = collection_task
	check FindOne(database, coll, bson.M{"proj": proj}, task)

	//insert
	err := InsertOne(database, collection_task,  bson.M{"$push":bson.M{"userids": user.UserId}}) 
	if err != nil {
		return errors.Wrap(err, "[database:"+ database + " coll:" + collection_task + " insert user: "+ user.UserId + " failed]")
	}
	return err  
}

