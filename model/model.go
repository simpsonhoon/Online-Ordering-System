package model

//model.go : db에 접속해 데이터를 핸들링, 결과 전달
import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	client       *mongo.Client
	colMenu      *mongo.Collection
	colOrderList *mongo.Collection
	colReview    *mongo.Collection
}

type OrderList struct {
	Menu       string `bson:"menu"`       //메뉴 이름
	Pnum       string `bson:"pnum"`       //고객 번호
	Address    string `bson:"address"`    //고객 주소
	OrderTime  string `bson:"orderTime"`  //주문 시간
	State      string `bson:"state"`      //주문 상태
	ChangeMenu string `bson:"changeMenu"` //주문 추가 및 변경 변수
}

type BurgerKing struct {
	Menu        string `bson:"menu"`        //메뉴이름
	Price       int    `bson:"price"`       // 가격
	Recommend   int    `bson:"recommend"`   //추천
	Grade       int    `bson:"grade"`       //평점
	ReleaseTime string `bson:"releaseTime"` //출시 시간
}

type MenuReview struct {
	Menu   string `bson:"menu"`   //메뉴이름
	Grade  int    `bson:"grade"`  //평점
	Review string `bson:"review"` //리뷰
}

// mongodb connect
func NewModel() (*Model, error) {
	r := &Model{}

	var err error
	mgUrl := "mongodb://127.0.0.1:27017"
	// Connect return *mongo.Client
	if r.client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mgUrl)); err != nil {
		return nil, err
	} else if err := r.client.Ping(context.Background(), nil); err != nil {
		return nil, err
	} else {
		db := r.client.Database("go-order")
		r.colMenu = db.Collection("menu-list")
		r.colOrderList = db.Collection("order-info")
		r.colReview = db.Collection("menu-review")
	}
	return r, nil
}

// 전체 메뉴 정렬 후 조회(주문자)

func (p *Model) GetAllMenu(sortOption string) []BurgerKing {

	filter := bson.D{}
	//높은 순으로 정렬 (평점 많은순, 최신순, 가격순)
	opts := options.Find().SetSort(bson.D{{sortOption, -1}})
	cursor, err := p.colMenu.Find(context.TODO(), filter, opts)
	var burgers []BurgerKing
	if err = cursor.All(context.TODO(), &burgers); err != nil {
		panic(err)
	}
	fmt.Println("Sorted by ", sortOption)
	for _, result := range burgers {
		//res, _ := json.Marshal(result)
		//fmt.Println(string(res))
		cursor.Decode(&result)
		output, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", output)
	}
	return burgers
}

// 메뉴이름으로 주문내역
func (p *Model) GetOrderListByMenu(flag, menuName string) (OrderList, error) {
	opts := []*options.FindOneOptions{}

	var filter bson.M
	if flag == "menu" {
		filter = bson.M{"menu": menuName}
	}

	var orderInfo OrderList
	if err := p.colOrderList.FindOne(context.TODO(), filter, opts...).Decode(&orderInfo); err != nil {
		return orderInfo, err
	} else {
		return orderInfo, nil
	}
}

func (p *Model) GetAllOrderList() []OrderList {
	filter := bson.D{}
	//높은 순으로 정렬 (평점 많은순, 최신순, 가격순)
	opts := options.Find().SetSort(bson.D{{"orderTime", -1}})
	cursor, err := p.colOrderList.Find(context.TODO(), filter, opts)
	var orders []OrderList
	if err = cursor.All(context.TODO(), &orders); err != nil {
		panic(err)
	}
	for _, result := range orders {
		//res, _ := json.Marshal(result)
		//fmt.Println(string(res))
		cursor.Decode(&result)
		output, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", output)
	}
	return orders

}

// 해당 메뉴에 대한 리뷰 및 평점 보기 (주문자)
func (p *Model) GetReview(menuName string) MenuReview {
	opts := []*options.FindOneOptions{}

	filter := bson.M{"menu": menuName}

	var review MenuReview
	if err := p.colReview.FindOne(context.TODO(), filter, opts...).Decode(&review); err != nil {
		panic(err)
	}
	return review
}

// 메뉴 주문
func (p *Model) OrderMenu(orderInfo OrderList) error {
	if _, err := p.colOrderList.InsertOne(context.TODO(), orderInfo); err != nil {
		fmt.Println("Your order failed")
		return fmt.Errorf(" Your order failed")
	}
	fmt.Println("Order Success")
	return nil
}

// 해당 메뉴의 리뷰 및 평점 작성
func (p *Model) WriteReview(review MenuReview) error {
	if _, err := p.colReview.InsertOne(context.TODO(), review); err != nil {
		fmt.Println("Failed to wirte review")
		return fmt.Errorf(" Failed to wirte review")
	}
	return nil
}

// 메뉴 업데이트 (주문자)
func (p *Model) ChangeMenu(menu, afterMenu string) error {
	filter := bson.M{"menu": menu}
	update := bson.M{
		"$set": bson.M{
			"menu": afterMenu,
		},
	}
	if _, err := p.colOrderList.UpdateOne(context.Background(), filter, update); err != nil {
		return err
	}
	return nil
}

//-----------------피주문자--------------------//

// 메뉴이름으로 조회 후 메뉴 정보 반환(피주문자)
func (p *Model) GetMenu(flag, menuName string) (BurgerKing, error) {
	opts := []*options.FindOneOptions{}

	var filter bson.M
	if flag == "menu" {
		filter = bson.M{"menu": menuName}
	}

	var burger BurgerKing
	if err := p.colMenu.FindOne(context.TODO(), filter, opts...).Decode(&burger); err != nil {
		return burger, err
	} else {
		return burger, nil
	}

}

// 메뉴 등록 (피주문자)
func (p *Model) CreateMenu(burger BurgerKing) error {
	if _, err := p.colMenu.InsertOne(context.TODO(), burger); err != nil {
		fmt.Println("Failed to create new menu")
		return fmt.Errorf(" Fail, create new menu")
	}
	return nil
}

// 메뉴 삭제 (피주문자)
func (p *Model) DeleteMenu(menuName string) error {
	filter := bson.M{"menu": menuName}

	if res, err := p.colMenu.DeleteOne(context.TODO(), filter); res.DeletedCount <= 0 {
		return fmt.Errorf("Could not Delete, There is no  %s in menu", menuName)
	} else if err != nil {
		return err
	}

	return nil
}

// 메뉴 업데이트 (피주문자)
func (p *Model) UpdateMenu(menuName string, price, recommend int) error {

	filter := bson.M{"menu": menuName}
	update := bson.M{
		"$set": bson.M{
			"price":     price,
			"recommend": recommend,
		},
	}

	if _, err := p.colMenu.UpdateOne(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}

// 주문 상태 업데이트(피주문자)
func (p *Model) UpdateState(menuName, state string) error {
	filter := bson.M{"menu": menuName}
	update := bson.M{
		"$set": bson.M{
			"state": state,
		},
	}
	if _, err := p.colOrderList.UpdateOne(context.Background(), filter, update); err != nil {
		return err
	}
	return nil
}
