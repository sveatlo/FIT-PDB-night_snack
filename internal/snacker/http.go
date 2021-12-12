package snacker

import (
	"github.com/gin-gonic/gin"
	cadre_http "github.com/moderntv/cadre/http"
	"github.com/moderntv/cadre/http/responses"
	"github.com/rs/zerolog"
	_ "google.golang.org/protobuf/types/known/structpb"

	"github.com/sveatlo/night_snack/internal/restaurant"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type HTTPGateway struct {
	log zerolog.Logger

	snackerSvc           *SnackerSvc
	restaurantCommandSvc *restaurant.CommandService
	restaurantQuerySvc   *restaurant.QueryService
}

func NewHTTP(snackerSvc *SnackerSvc, restaurantCommandSvc *restaurant.CommandService, restaurantQuerySvc *restaurant.QueryService, log zerolog.Logger) (g *HTTPGateway, err error) {
	g = &HTTPGateway{
		log: log.With().Str("component", "http").Logger(),

		snackerSvc:           snackerSvc,
		restaurantCommandSvc: restaurantCommandSvc,
		restaurantQuerySvc:   restaurantQuerySvc,
	}

	return
}

func (gw *HTTPGateway) GetRoutes() cadre_http.RoutingGroup {
	return cadre_http.RoutingGroup{
		Base:       "",
		Middleware: []gin.HandlerFunc{},
		Routes:     map[string]map[string][]gin.HandlerFunc{},
		Groups: []cadre_http.RoutingGroup{
			{
				Base:       "/restaurant",
				Middleware: []gin.HandlerFunc{},
				Routes: map[string]map[string][]gin.HandlerFunc{
					"/": {
						"GET":    {gw.getRestaurants},
						"POST":   {gw.createRestaurant},
						"PUT":    {gw.updateRestaurant},
						"DELETE": {gw.deleteRestaurant},
					},
					"/:restaurant_id": {
						"GET":    {gw.getRestaurant},
						"PUT":    {gw.updateRestaurant},
						"DELETE": {gw.deleteRestaurant},
					},
				},
				Groups: []cadre_http.RoutingGroup{
					{
						Base: "/:restaurant_id",
						Routes: map[string]map[string][]gin.HandlerFunc{
							"/category": {
								"POST": {gw.createMenuCategory},
							},
						},
					},
				},
			},
		},
	}
}

// getRestaurants
// @Summary Gets restaurants
// @Description Get all restaurants
// @ID restaurants_get
// @Router /restaurant/ [get]
// @Success 200      {object} responses.SuccessResponse{data=[]restaurant_pb.Restaurant}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) getRestaurants(c *gin.Context) {
	restaurants, err := gw.restaurantQuerySvc.GetAll(c.Request.Context(), &restaurant_pb.GetRestaurants{})
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, restaurants.Restaurants)
}

// getRestaurant
// @Summary Gets restaurants
// @Description Get all restaurants
// @ID restaurants_get
// @Router /restaurant/{restaurant_id} [get]
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.Restaurant}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) getRestaurant(c *gin.Context) {
	id := c.Param("restaurant_id")

	restaurant, err := gw.restaurantQuerySvc.Get(c.Request.Context(), &restaurant_pb.GetRestaurant{Id: id})
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, restaurant)
}

// createRestaurant
// @Summary Creates restaurant
// @Description Create new restaurant
// @ID restaurant_create
// @Router /restaurant/ [post]
// @Param   createRestaurantCmd body restaurant_pb.CreateRestaurant true "CreateRestaurant command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.RestaurantCreated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) createRestaurant(c *gin.Context) {
	createRestaurantCmd := &restaurant_pb.CmdRestaurantCreate{}
	if err := c.Bind(&createRestaurantCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}

	res, err := gw.restaurantCommandSvc.Create(c.Request.Context(), createRestaurantCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// updateRestaurant
// @Summary Updates restaurant
// @Description Update existing restaurant. This can be called either with or without the ID in URL, however the ID in URL takes precedence
// @ID restaurant_update
// @Router /restaurant/{restaurant_id} [put]
// @Param   updateRestaurantCmd body restaurant_pb.UpdateRestaurant true "UpdateRestaurant command data"
// @Success 200      {object} responses.SuccessResponse{}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) updateRestaurant(c *gin.Context) {
	updateRestaurantCmd := &restaurant_pb.CmdRestaurantUpdate{}
	if err := c.Bind(&updateRestaurantCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	id := c.Param("restaurant_id")
	if id != "" {
		updateRestaurantCmd.Id = id
	}

	res, err := gw.restaurantCommandSvc.Update(c.Request.Context(), updateRestaurantCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// deleteRestaurant
// @Summary Deletes restaurant
// @Description Delete new restaurant
// @ID restaurant_delete
// @Router /restaurant/{restaurant_id} [put]
// @Param   deleteRestaurantCmd body restaurant_pb.DeleteRestaurant true "DeleteRestaurant command data"
// @Success 200      {object} responses.SuccessResponse{}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) deleteRestaurant(c *gin.Context) {
	deleteRestaurantCmd := &restaurant_pb.CmdRestaurantDelete{}
	if err := c.Bind(&deleteRestaurantCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	id := c.Param("restaurant_id")
	if id != "" {
		deleteRestaurantCmd.Id = id
	}

	res, err := gw.restaurantCommandSvc.Delete(c.Request.Context(), deleteRestaurantCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// createMenuCategory
// @Summary Creates menu category in restaurant
// @Description Create new restaurant
// @ID restaurant_create
// @Router /restaurant/{restaurant_id}/category [post]
// @Param   createRestaurantCmd body restaurant_pb.CreateRestaurant true "CreateRestaurant command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.RestaurantCreated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) createMenuCategory(c *gin.Context) {
	createRestaurantCmd := &restaurant_pb.CmdRestaurantCreate{}
	if err := c.Bind(&createRestaurantCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}

	res, err := gw.restaurantCommandSvc.Create(c.Request.Context(), createRestaurantCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}
