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
						Base: "/:restaurant_id/menu_categories",
						Routes: map[string]map[string][]gin.HandlerFunc{
							"/": {
								"POST": {gw.createMenuCategory},
							},
							"/:menu_category_id": {
								"PUT":    {gw.updateMenuCategory},
								"DELETE": {gw.deleteMenuCategory},
							},
						},
						Groups: []cadre_http.RoutingGroup{
							{
								Base: "/:menu_category_id/items",
								Routes: map[string]map[string][]gin.HandlerFunc{
									"/": {
										"POST": {gw.createMenuItem},
									},
									"/:menu_item_id": {
										"PUT":    {gw.updateMenuItem},
										"DELETE": {gw.deleteMenuItem},
									},
								},
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
// @Param   cmd body restaurant_pb.CmdRestaurantCreate true "Command data"
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
// @Param   cmd body restaurant_pb.CmdRestaurantUpdate true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.RestaurantUpdated}
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
// @Description Delete an existing restaurant
// @ID restaurant_delete
// @Router /restaurant/{restaurant_id} [delete]
// @Param   cmd body restaurant_pb.CmdRestaurantDelete true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.RestaurantDeleted}
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
// @Summary Create menu category
// @Description Creates menu category in restaurant
// @ID restaurant_create
// @Router /restaurant/{restaurant_id}/menu_categories [post]
// @Param   cmd body restaurant_pb.CmdMenuCategoryCreate true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuCategoryCreated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) createMenuCategory(c *gin.Context) {
	createMenuCategoryCmd := &restaurant_pb.CmdMenuCategoryCreate{}
	if err := c.Bind(&createMenuCategoryCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	restaurantID := c.Param("restaurant_id")
	if restaurantID != "" {
		createMenuCategoryCmd.RestaurantId = restaurantID
	}

	res, err := gw.restaurantCommandSvc.CreateMenuCategory(c.Request.Context(), createMenuCategoryCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// updateMenuCategory
// @Summary Update menu category
// @Description Update menu category in restaurant
// @ID restaurant_update
// @Router /restaurant/{restaurant_id}/menu_categories/{menu_category_id} [put]
// @Param   cmd body restaurant_pb.CmdMenuCategoryUpdate true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuCategoryUpdated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) updateMenuCategory(c *gin.Context) {
	updateMenuCategoryCmd := &restaurant_pb.CmdMenuCategoryUpdate{}
	if err := c.Bind(&updateMenuCategoryCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	menuCategoryID := c.Param("menu_category_id")
	if menuCategoryID != "" {
		updateMenuCategoryCmd.Id = menuCategoryID
	}

	res, err := gw.restaurantCommandSvc.UpdateMenuCategory(c.Request.Context(), updateMenuCategoryCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// deleteMenuCategory
// @Summary Delete menu category
// @Description Delete menu category in restaurant
// @ID restaurant_delete
// @Router /restaurant/{restaurant_id}/menu_categories/{menu_category_id} [delete]
// @Param   cmd body restaurant_pb.CmdMenuCategoryDelete true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuCategoryDeleted}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) deleteMenuCategory(c *gin.Context) {
	deleteMenuCategoryCmd := &restaurant_pb.CmdMenuCategoryDelete{}
	if err := c.Bind(&deleteMenuCategoryCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	menuCategoryID := c.Param("menu_category_id")
	if menuCategoryID != "" {
		deleteMenuCategoryCmd.Id = menuCategoryID
	}

	res, err := gw.restaurantCommandSvc.DeleteMenuCategory(c.Request.Context(), deleteMenuCategoryCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// createMenuItem
// @Summary Create menu item
// @Description Creates menu item in restaurant
// @ID menu_item_create
// @Router /restaurant/{restaurant_id}/menu_categories [post]
// @Param   cmd body restaurant_pb.CmdMenuItemCreate true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuItemCreated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) createMenuItem(c *gin.Context) {
	createMenuItemCmd := &restaurant_pb.CmdMenuItemCreate{}
	if err := c.Bind(&createMenuItemCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	restaurantID := c.Param("restaurant_id")
	if restaurantID != "" {
		createMenuItemCmd.RestaurantId = restaurantID
	}
	categoryID := c.Param("menu_category_id")
	if categoryID != "" {
		createMenuItemCmd.CategoryId = categoryID
	}

	res, err := gw.restaurantCommandSvc.CreateMenuItem(c.Request.Context(), createMenuItemCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// updateMenuItem
// @Summary Update menu item
// @Description Update menu item in restaurant
// @ID menu_item_update
// @Router /restaurant/{restaurant_id}/menu_categories/{menu_item_id} [put]
// @Param   cmd body restaurant_pb.CmdMenuItemUpdate true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuItemUpdated}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) updateMenuItem(c *gin.Context) {
	updateMenuItemCmd := &restaurant_pb.CmdMenuItemUpdate{}
	if err := c.Bind(&updateMenuItemCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	restaurantID := c.Param("restaurant_id")
	if restaurantID != "" {
		updateMenuItemCmd.RestaurantId = restaurantID
	}
	categoryID := c.Param("menu_category_id")
	if categoryID != "" {
		updateMenuItemCmd.CategoryId = categoryID
	}
	itemID := c.Param("menu_item_id")
	if itemID != "" {
		updateMenuItemCmd.Id = itemID
	}

	res, err := gw.restaurantCommandSvc.UpdateMenuItem(c.Request.Context(), updateMenuItemCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}

// deleteMenuItem
// @Summary Delete menu item
// @Description Delete menu item in restaurant
// @ID menu_item_delete
// @Router /restaurant/{restaurant_id}/menu_categories/{menu_item_id} [put]
// @Param   cmd body restaurant_pb.CmdMenuItemDelete true "Command data"
// @Success 200      {object} responses.SuccessResponse{data=restaurant_pb.MenuItemDeleted}
// @Failure 400,500  {object} responses.ErrorResponse
func (gw *HTTPGateway) deleteMenuItem(c *gin.Context) {
	deleteMenuItemCmd := &restaurant_pb.CmdMenuItemDelete{}
	if err := c.Bind(&deleteMenuItemCmd); err != nil {
		responses.BadRequest(c, responses.NewError(err))
		return
	}
	restaurantID := c.Param("restaurant_id")
	if restaurantID != "" {
		deleteMenuItemCmd.RestaurantId = restaurantID
	}
	itemID := c.Param("menu_item_id")
	if itemID != "" {
		deleteMenuItemCmd.Id = itemID
	}

	res, err := gw.restaurantCommandSvc.DeleteMenuItem(c.Request.Context(), deleteMenuItemCmd)
	if err != nil {
		responses.InternalError(c, responses.NewError(err))
		return
	}

	responses.Ok(c, res)
}
