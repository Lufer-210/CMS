package router

import (
	"CMS/internal/handler/admin"
	"CMS/internal/handler/block"
	"CMS/internal/handler/post"
	"CMS/internal/handler/user"
	"CMS/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	// 全局错误处理中间件
	r.Use(middleware.GlobalErrorHandler())

	const pre = "/api"

	// 公开路由
	public := r.Group(pre)
	{
		public.POST("/user/reg", user.Register) // 用户注册
		public.POST("/user/login", user.Login)  // 用户登录
	}

	// 需要身份验证的基础路由组
	auth := r.Group(pre)
	auth.Use(middleware.JWTAuthMiddleware())
	{
		// 学生路由
		student := auth.Group("/student")
		{
			student.GET("/post", post.GetAllPosts)           // 获取所有帖子
			student.POST("/post", post.CreatePost)           // 发布帖子
			student.DELETE("/post", post.DeletePost)         // 删除帖子
			student.POST("/report-post", block.ReportPost)   // 举报帖子
			student.PUT("/post", post.UpdatePost)            // 修改帖子
			student.GET("/likes", post.GetPostLikes)         // 获取帖子点赞数
			student.GET("/report-post", block.GetReportList) // 查看举报审批
			student.POST("/likes", post.LikePost)            // 点赞帖子
		}

		// 管理员路由 - 需要额外的管理员权限验证
		adminGroup := auth.Group("/admin")
		adminGroup.Use(middleware.AdminAuthMiddleware())
		{
			adminGroup.GET("/report", admin.GetPendingReports) // 获取待审批举报
			adminGroup.POST("/report", admin.ApproveReport)    // 审批举报
		}
	}
}
