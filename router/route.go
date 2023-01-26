package router

//router.go : api 전체 인입에 대한 관리 및 구성을 담당하는 파일
import (
	"fmt"
	ctl "lecture/oos/controller"
	"lecture/oos/docs" //swagger에 의해 자동 생성된 package
	"lecture/oos/logger"

	"github.com/gin-gonic/gin"
	swgFiles "github.com/swaggo/files"
	ginSwg "github.com/swaggo/gin-swagger"
)

type Router struct {
	ct *ctl.Controller
}

func NewRouter(ctl *ctl.Controller) (*Router, error) {
	r := &Router{ct: ctl} //controller 포인터를 ct로 복사, 할당
	return r, nil
}

// cross domain을 위해 사용
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		//허용할 header 타입에 대해 열거
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Forwarded-For, Authorization, accept, origin, Cache-Control, X-Requested-With")
		//허용할 method에 대해 열거
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// 임의 인증을 위한 함수
func liteAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c == nil {
			c.Abort() // 미들웨어에서 사용, 이후 요청 중지
			return
		}
		//http 헤더내 "Authorization" 폼의 데이터를 조회
		auth := c.GetHeader("Authorization")
		//실제 인증기능이 올수있다. 단순히 출력기능만 처리 현재는 출력예시
		fmt.Println("Authorization-word ", auth)

		c.Next() // 다음 요청 진행
	}
}

// 실제 라우팅
func (p *Router) Idx() *gin.Engine {
	e := gin.New() // gin선언

	e.Use(logger.GinLogger())       // gin 내부 log, logger 미들웨어 사용 선언
	e.Use(logger.GinRecovery(true)) //gin 내부 recover, recovery 미들웨어 사용 - 패닉복구
	e.Use(CORS())                   // crossdomain 미들웨어 사용

	logger.Info("start server")
	e.GET("/swagger/:any", ginSwg.WrapHandler(swgFiles.Handler))
	docs.SwaggerInfo.Host = "localhost" //swagger 정보 등록

	customer := e.Group("/customer", liteAuth())
	{
		fmt.Println(customer)
		customer.GET("/getMenu/:sortOption", p.ct.GetMenu)   //메뉴 리스트 출력 조회
		customer.GET("/getReview/:menuName", p.ct.GetReview) //메뉴별 평점 및 리뷰 조회
		customer.POST("/writeReview", p.ct.WriteReview)      //메뉴별 평점 작성
		customer.POST("orderMenu", p.ct.OrderMenu)           //메뉴 선택 후 주문
		customer.PUT("changeMenu", p.ct.ChangeMenu)          // 메뉴변경
		customer.PUT("addMenu", p.ct.AddMenu)                //메뉴 추가
		customer.GET("getOrderState", p.ct.GetAllOrderList)  //주문 내역(상태) 조회
	}

	seller := e.Group("/seller", liteAuth())
	{
		fmt.Println(seller)
		seller.PUT("/updateMenu", p.ct.UpdateMenu)             //메뉴 수정
		seller.POST("/register", p.ct.RegisterMenu)            //신규메뉴 등록
		seller.PUT("/updateOrderState", p.ct.UpdateOrderState) //주문내역 조회 및 상태 변경
		seller.DELETE("/delete/:menu", p.ct.DeleteMenu)        //메뉴 삭제
	}

	return e
}
