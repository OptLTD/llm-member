// 页面管理
const PageManager = {
  currentPage: "home",

  // 初始化页面
  init() {
    this.handleHashChange();
    window.addEventListener("hashchange", () => this.handleHashChange());
  },

  // 处理路由变化
  handleHashChange() {
    const hash = window.location.hash.slice(1) || "home";
    this.showPage(hash);
  },

  // 显示页面
  showPage(pageName) {
    this.currentPage = pageName;

    // 这里可以根据页面名称加载不同的内容
    switch (pageName) {
      case "profile":
        this.showProfilePage();
        break;
      case "usage":
        this.showUsagePage();
        break;
      case "api-keys":
        this.showApiKeysPage();
        break;
      default:
        // 默认显示首页内容
        break;
    }
  },

  // 显示个人资料页面
  showProfilePage() {
    if (!Auth.isLoggedIn) {
      Utils.showNotification("请先登录", "warning");
      return;
    }

    // 这里可以创建个人资料页面的内容
    console.log("显示个人资料页面");
  },

  // 显示使用统计页面
  showUsagePage() {
    if (!Auth.isLoggedIn) {
      Utils.showNotification("请先登录", "warning");
      return;
    }

    // 这里可以创建使用统计页面的内容
    console.log("显示使用统计页面");
  },

  // 显示API密钥页面
  showApiKeysPage() {
    if (!Auth.isLoggedIn) {
      Utils.showNotification("请先登录", "warning");
      return;
    }

    // 这里可以创建API密钥页面的内容
    console.log("显示API密钥页面");
  },
};

// 页面加载完成后初始化
document.addEventListener("DOMContentLoaded", () => {
  PageManager.init();

  // 设置事件监听器
  if (window.App && window.App.Pages) {
    window.App.Pages.setupPageClick();
  }

  // 检查用户登录状态并更新UI
  if (window.App && window.App.Auth) {
    window.App.Auth.checkAuth();
  }
});
