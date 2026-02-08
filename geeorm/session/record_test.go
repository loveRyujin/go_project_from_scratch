package session

import (
	"testing"
)

func testRecordInit(t *testing.T) (*Session, func()) {
	t.Helper()
	s, cleanup := newTestSession(t)
	s.Model(&User{})
	_ = s.DropTable()
	err := s.CreateTable()
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return s, cleanup
}

func TestSession_Insert(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 测试插入单条记录
	affected, err := s.Insert(&User{"Tom", 18})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestSession_Insert_Multiple(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 测试批量插入多条记录
	user1 := &User{"Tom", 18}
	user2 := &User{"Sam", 25}
	user3 := &User{"Jack", 30}

	affected, err := s.Insert(user1, user2, user3)
	if err != nil {
		t.Fatalf("failed to insert multiple: %v", err)
	}
	if affected != 3 {
		t.Errorf("expected 3 rows affected, got %d", affected)
	}
}

func TestSession_Find(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 先插入测试数据
	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25})

	// 查询所有记录
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestSession_Find_VerifyData(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 插入已知数据
	_, _ = s.Insert(&User{"Alice", 20})

	// 查询并验证数据正确性
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[0].Age != 20 {
		t.Errorf("expected {Alice, 20}, got {%s, %d}", users[0].Name, users[0].Age)
	}
}

func TestSession_Find_Empty(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 空表查询
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find from empty table: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestSession_Find_InvalidParam(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 传入非指针应该返回错误
	var users []User
	err := s.Find(users)
	if err == nil {
		t.Error("expected error when passing non-pointer to Find")
	}

	// 传入非切片指针应该返回错误
	var user User
	err = s.Find(&user)
	if err == nil {
		t.Error("expected error when passing non-slice pointer to Find")
	}

	// 传入非结构体切片应该返回错误
	var names []string
	err = s.Find(&names)
	if err == nil {
		t.Error("expected error when passing non-struct slice to Find")
	}
}

func TestSession_Update_WithMap(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18})

	// 使用 map 更新
	affected, err := s.Where("Name = ?", "Tom").Update(map[string]any{"Age": 25})
	if err != nil {
		t.Fatalf("failed to update with map: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestSession_Update_WithKV(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18})

	// 使用 kv 平铺更新
	affected, err := s.Where("Name = ?", "Tom").Update("Age", 30)
	if err != nil {
		t.Fatalf("failed to update with kv: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestSession_Update_WithStruct(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18})

	// 使用 struct 更新
	affected, err := s.Where("Name = ?", "Tom").Update(User{Name: "Tom", Age: 28})
	if err != nil {
		t.Fatalf("failed to update with struct: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// 验证数据已更新
	var users []User
	_ = s.Find(&users)
	if len(users) != 1 || users[0].Age != 28 {
		t.Errorf("expected age 28 after update, got %v", users)
	}
}

func TestSession_Update_InvalidParam(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 无参数
	_, err := s.Update()
	if err == nil {
		t.Error("expected error for empty Update args")
	}

	// 奇数个 kv 参数
	_, err = s.Update("Name", "Tom", "Age")
	if err == nil {
		t.Error("expected error for odd kv args")
	}

	// kv 的 key 不是 string
	_, err = s.Update(123, "Tom")
	if err == nil {
		t.Error("expected error for non-string kv key")
	}

	// 单参数传入非 struct/map
	_, err = s.Update("just a string")
	if err == nil {
		t.Error("expected error for non-struct/map single arg")
	}
}

func TestSession_Delete(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25})

	// 按条件删除
	affected, err := s.Where("Name = ?", "Tom").Delete()
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// 验证只剩一条
	count, _ := s.Count()
	if count != 1 {
		t.Errorf("expected 1 remaining record, got %d", count)
	}
}

func TestSession_Count(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 空表计数
	count, err := s.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// 插入数据后计数
	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25}, &User{"Jack", 30})
	count, err = s.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestSession_Limit(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25}, &User{"Jack", 30})

	var users []User
	err := s.Limit(2).Find(&users)
	if err != nil {
		t.Fatalf("failed to find with limit: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestSession_Where(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25}, &User{"Jack", 30})

	var users []User
	err := s.Where("Age > ?", 20).Find(&users)
	if err != nil {
		t.Fatalf("failed to find with where: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users (Age > 20), got %d", len(users))
	}
}

func TestSession_OrderBy(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25}, &User{"Jack", 30})

	// 升序
	var users []User
	err := s.OrderBy("Age", false).Find(&users)
	if err != nil {
		t.Fatalf("failed to find with order: %v", err)
	}
	if users[0].Name != "Tom" || users[2].Name != "Jack" {
		t.Errorf("expected ascending order, got %v", users)
	}

	// 降序
	var usersDesc []User
	err = s.OrderBy("Age", true).Find(&usersDesc)
	if err != nil {
		t.Fatalf("failed to find with desc order: %v", err)
	}
	if usersDesc[0].Name != "Jack" || usersDesc[2].Name != "Tom" {
		t.Errorf("expected descending order, got %v", usersDesc)
	}
}

func TestSession_ChainCall(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25}, &User{"Jack", 30}, &User{"Alice", 22})

	// 链式调用: Where + OrderBy + Limit
	var users []User
	err := s.Where("Age > ?", 18).OrderBy("Age", true).Limit(2).Find(&users)
	if err != nil {
		t.Fatalf("failed chain call: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Jack" {
		t.Errorf("expected first user Jack (oldest), got %s", users[0].Name)
	}
}

func TestSession_First(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25})

	// 查询第一条记录
	var user User
	err := s.OrderBy("Age", false).First(&user)
	if err != nil {
		t.Fatalf("failed to get first: %v", err)
	}
	if user.Name != "Tom" || user.Age != 18 {
		t.Errorf("expected {Tom, 18}, got {%s, %d}", user.Name, user.Age)
	}
}

func TestSession_First_WithWhere(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25})

	// 按条件查询第一条
	var user User
	err := s.Where("Name = ?", "Sam").First(&user)
	if err != nil {
		t.Fatalf("failed to get first with where: %v", err)
	}
	if user.Name != "Sam" || user.Age != 25 {
		t.Errorf("expected {Sam, 25}, got {%s, %d}", user.Name, user.Age)
	}
}

func TestSession_First_NotFound(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 空表查询应返回 "record not found"
	var user User
	err := s.First(&user)
	if err == nil {
		t.Fatal("expected error for empty table")
	}
	if err.Error() != "record not found" {
		t.Errorf("expected 'record not found', got '%s'", err.Error())
	}
}

func TestSession_Insert_InvalidParam(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 传入非结构体应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when passing non-struct to Insert")
		}
	}()

	s.Insert("not a struct")
}
