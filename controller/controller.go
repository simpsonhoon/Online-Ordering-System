package controller

// /controller.go : 실제 비지니스 로직 및 프로세스가 처리후 결과 전송
import (
	"encoding/json"
	"fmt"
	"lecture/oos/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	md *model.Model
}

func NewCTL(rep *model.Model) (*Controller, error) {
	r := &Controller{md: rep}
	return r, nil
}

func (p *Controller) RespOK(c *gin.Context, resp interface{}) {
	c.JSON(http.StatusOK, resp)
}

func (p *Controller) RespError(c *gin.Context, body interface{}, status int, err ...interface{}) {
	bytes, _ := json.Marshal(body)
	fmt.Println("Request error", "path", c.FullPath(), "body", bytes, "status", status, "error", err)

	c.JSON(status, gin.H{
		"Error":  "Request Error",
		"path":   c.FullPath(),
		"body":   bytes,
		"status": status,
		"error":  err,
	})
	c.Abort()
}

// ----------주문자--------------------------//
// 메뉴 리스트 출력 조회 (주문자)

// GetMenu godoc
// @Summary call GetMenu, return sortOption, BurgerKing menu by json.
// @Description 메뉴 리스트의 정렬 기준을 정하고 조회기능(주문자가 수행)
// @name GetMenu
// @Accept  json
// @Produce  json
// @Param sortOption path string true "sortOption"
// @Router /customer/getMenu/:sortOption [get]
// @Success 200 {object} Controller
func (p *Controller) GetMenu(c *gin.Context) {
	r, _ := model.NewModel()
	sortOption := c.Param("sortOption")
	c.JSON(200, gin.H{"Sort Option": sortOption, "Menu List": r.GetAllMenu(sortOption)})
	c.Next()
}

// GetReview godoc
// @Summary call GetReview, return MenuReview by json.
// @Description 메뉴별 평점 및 리뷰 조회기능(주문자가 수행)
// @name GetReview
// @Accept  json
// @Produce  json
// @Param menuName path string true "menuName"
// @Router /customer/getReview/:menuName [get]
// @Success 200 {object} Controller
func (p *Controller) GetReview(c *gin.Context) {
	r, _ := model.NewModel()
	menuName := c.Param("menuName")
	review := r.GetReview(menuName)
	if review == (model.MenuReview{}) { //해당 메뉴의 리뷰 내역이 없으면
		p.RespError(c, nil, http.StatusUnprocessableEntity, "You didn`t wirte review before", nil)
		return
	}

	c.JSON(200, review)
	c.Next()
}

// WriteReview godoc
// @Summary call WriteReview, return "Your review registered" by json.
// @Description 메뉴별 평점 작성기능(주문자가 수행)
// @name WriteReview
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param grade path string true "grade"
// @Param review path string true "review"
// @Router /customer/writeReview [post]
// @Success 200 {object} Controller
func (p *Controller) WriteReview(c *gin.Context) {
	menuName := c.PostForm("menu")
	sGrade := c.PostForm("grade")
	review := c.PostForm("review")

	if len(menuName) <= 0 || len(review) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	orderList, _ := p.md.GetOrderListByMenu("menu", menuName)
	if orderList == (model.OrderList{}) { //해당 메뉴의 주문 내역이 없으면
		p.RespError(c, nil, http.StatusUnprocessableEntity, "You didn`t ordered that menu before", nil)
		return
	}

	grade, _ := strconv.Atoi(sGrade)
	req := model.MenuReview{Menu: menuName, Grade: grade, Review: review} //리뷰 db에 저장
	if err := p.md.WriteReview(req); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "You didn`t order that menu", nil)
		return
	}

	c.JSON(200, gin.H{"result": "Your review registered"})
	c.Next()

}

// OrderMenu godoc
// @Summary call OrderMenu, return "Order Success", count by json.
// @Description 메뉴 주문기능과 주문번호 받는 기능(주문자가 수행)
// @name OrderMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param pnum path string true "pnum"
// @Param address path string true "address"
// @Router /customer/orderMenu [post]
// @Success 200 {object} Controller
func (p *Controller) OrderMenu(c *gin.Context) {
	menuName := c.PostForm("menu")
	pnum := c.PostForm("pnum")
	address := c.PostForm("address")
	orderTime := time.Now().Format("2006-01-02 15:04:05")
	state := "접수중" //최초 상태는 접수중...

	if len(menuName) <= 0 || len(address) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	req := model.OrderList{Menu: menuName, Pnum: pnum, Address: address, OrderTime: orderTime, State: state}

	if err := p.md.OrderMenu(req); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	r, _ := model.NewModel()
	count := 0
	orders := r.GetAllOrderList()
	for _, order := range orders {
		if order == (model.OrderList{}) { //주문 내역 없으면 null
			p.RespError(c, nil, http.StatusUnprocessableEntity, "Null", nil)
			return
		}
		count++
	}

	c.JSON(200, gin.H{
		"result":       "Order Success",
		"Order Number": count, //주문번호
	})
	c.Next()
}

// AddMenu godoc
// @Summary call AddMenu, return success,fail by json.
// @Description 메뉴추가 기능과 배달중이면 신규주문 접수 기능(주문자가 수행)
// @name AddMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param changeMenu path string true "changeMenu"
// @Router /customer/addMenu [put]
// @Success 200 {object} Controller
func (p *Controller) AddMenu(c *gin.Context) {
	beforeMenu := c.PostForm("menu")
	addMenu := c.PostForm("changeMenu")

	orderList, _ := p.md.GetOrderListByMenu("menu", beforeMenu)
	if orderList == (model.OrderList{}) { //해당 메뉴의 주문 내역이 없으면
		p.RespError(c, nil, http.StatusUnprocessableEntity, " You didn`t ordered that menu before", nil)
		return
	}

	if orderList.State == "배달중" { // 신규 주문으로 전환

		pnum := orderList.Pnum
		address := orderList.Address
		orderTime := time.Now().Format("2006-01-02 15:04:05")
		state := "접수중"
		req := model.OrderList{Menu: addMenu, Pnum: pnum, Address: address, OrderTime: orderTime, State: state}

		if err := p.md.OrderMenu(req); err != nil {
			p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
			return
		}
		c.JSON(200, gin.H{
			"msg":       "Sorry, You can not add menu.I will make you new order",
			"New order": req,
		})
		c.Next()
	} else {
		addMenu = beforeMenu + " , " + addMenu
		if err := p.md.ChangeMenu(beforeMenu, addMenu); err != nil {
			p.RespError(c, nil, http.StatusUnprocessableEntity, "Fail,parameter not found", nil)
			return
		}
		c.JSON(200, gin.H{"msg": "Menu add success"})
		c.Next()
	}

}

// ChangeMenu godoc
// @Summary call ChangeMenu, return success,fail by json.
// @Description 메뉴변경 기능과 조리중/배달중이면 변경 미수행 기능(주문자가 수행)
// @name ChangeMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param changeMenu path string true "changeMenu"
// @Router /customer/changeMenu [put]
// @Success 200 {object} Controller
func (p *Controller) ChangeMenu(c *gin.Context) {
	beforeMenu := c.PostForm("menu")
	afterMenu := c.PostForm("changeMenu")

	if len(beforeMenu) <= 0 || len(afterMenu) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	orderList, _ := p.md.GetOrderListByMenu("menu", beforeMenu)
	if orderList == (model.OrderList{}) { //해당 메뉴의 주문 내역이 없으면
		p.RespError(c, nil, http.StatusUnprocessableEntity, " You didn`t ordered that menu before", nil)
		return
	}

	if orderList.State == "조리중" || orderList.State == "배달중" {
		c.JSON(200, gin.H{"msg": "Sorry, You can not change menu."})
		c.Next()
	} else if orderList.State == "접수중" {
		if err := p.md.ChangeMenu(beforeMenu, afterMenu); err != nil {
			p.RespError(c, nil, http.StatusUnprocessableEntity, "Fail,parameter not found", nil)
			return
		}
		c.JSON(200, gin.H{"msg": " Menu change success"})
		c.Next()
	}
	c.Next()
}

// GetAllOrderList godoc
// @Summary call GetAllOrderList, return OrderList by json.
// @Description 전체 주문 내역 조회(주문자 수행)
// @name GetAllOrderList
// @Accept  json
// @Produce  json
// @Router /customer/getAllOrderList [get]
// @Success 200 {object} Controller
func (p *Controller) GetAllOrderList(c *gin.Context) {
	r, _ := model.NewModel()

	orders := r.GetAllOrderList()
	for _, order := range orders {
		if order == (model.OrderList{}) { //주문 내역 없으면 null
			p.RespError(c, nil, http.StatusUnprocessableEntity, "Null", nil)
			return
		}
	}
	c.JSON(200, gin.H{"Menu List": orders})
	c.Next()
}

// ---------------피주문자--------------------

// UpdateMenu godoc
// @Summary call UpdateMenu, return "Menu change success" by json.
// @Description 메뉴판 수정 기능(피주문자가 수행)
// @name UpdateMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param price path string true "price"
// @Param recommend path string true "recommend"
// @Router /seller/updateMenu [put]
// @Success 200 {object} Controller
func (p *Controller) UpdateMenu(c *gin.Context) {
	menuName := c.PostForm("menu")
	sPrice := c.PostForm("price")
	sRecommend := c.PostForm("recommend")

	if len(menuName) <= 0 || len(sPrice) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	burger, _ := p.md.GetMenu("menu", menuName) //메뉴이름으로 메뉴 정보 가져오기
	if burger == (model.BurgerKing{}) {
		p.RespError(c, nil, http.StatusUnprocessableEntity, " Can`t find that menu", nil)
		return
	}

	price, _ := strconv.Atoi(sPrice)
	recommend, _ := strconv.Atoi(sRecommend)

	if err := p.md.UpdateMenu(menuName, price, recommend); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	c.JSON(200, gin.H{"msg": "Menu change success"})
	c.Next()
}

// DeleteMenu godoc
// @Summary call DeleteMenu, return "Delete menu success" by json.
// @Description 메뉴판 삭제 기능(피주문자가 수행)
// @name DeleteMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Router /seller/delete/:menu [delete]
// @Success 200 {object} Controller
func (p *Controller) DeleteMenu(c *gin.Context) {
	menuName := c.Param("menu")

	if len(menuName) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	if err := p.md.DeleteMenu(menuName); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "Menu delete Fail!", nil)
		return
	}

	c.JSON(200, gin.H{"result": "Delete menu success"})
	c.Next()

}

// RegisterMenu godoc
// @Summary call RegisterMenu, return ""Register menu Success" by json.
// @Description 신규메뉴 등록기능(피주문자가 수행)
// @name RegisterMenu
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param price path string true "price"
// @Param recommend path string true "recommend"
// @Router /seller/register [post]
// @Success 200 {object} Controller
func (p *Controller) RegisterMenu(c *gin.Context) {

	menuName := c.PostForm("menu")
	sPrice := c.PostForm("price")
	sRecommend := c.PostForm("recommend")

	if len(menuName) <= 0 || len(sPrice) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	price, _ := strconv.Atoi(sPrice)
	recommend, _ := strconv.Atoi(sRecommend)
	grade := 0 //최초 평점은 0점
	releaseTime := time.Now().Format("2006-01-02 15:04:05")

	req := model.BurgerKing{Menu: menuName, Price: price, Recommend: recommend, Grade: grade, ReleaseTime: releaseTime}

	if err := p.md.CreateMenu(req); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "parameter not found", nil)
		return
	}

	c.JSON(200, gin.H{"result": "Register menu Success"})
	c.Next()
}

// UpdateOrderState godoc
// @Summary call UpdateOrderState, return "State change success" by json.
// @Description 주문내역 조회 및 상태 변경(피주문자가 수행)
// @name UpdateOrderState
// @Accept  json
// @Produce  json
// @Param menu path string true "menu"
// @Param state path string true "state"
// @Router /seller/updateOrderState [put]
// @Success 200 {object} Controller
func (p *Controller) UpdateOrderState(c *gin.Context) {
	menuName := c.PostForm("menu")
	state := c.PostForm("state")
	if len(menuName) <= 0 || len(state) <= 0 {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "Fail,parameter not found", nil)
		return
	}

	if err := p.md.UpdateState(menuName, state); err != nil {
		p.RespError(c, nil, http.StatusUnprocessableEntity, "Fail,parameter not found", nil)
		return
	}

	fmt.Println("State changed")
	c.JSON(200, gin.H{"msg": "State change success", menuName: state})
	c.Next()
}
