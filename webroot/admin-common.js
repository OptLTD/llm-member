class LLMMemberApp {
  constructor() {
    this.token = localStorage.getItem("token");
    this.apiKey = localStorage.getItem("apiKey");
    this.currentPage = "dashboard";
    this.allModels = []; // 存储所有模型数据
    this.init();
  }

  init() {
    this.bindEvents();
    this.checkAuth();
    // 初始化模块
    this.dashboardManager = new DashboardManager(this);
    this.testChatManager = new TestChatManager(this);
    this.pricingManager = new PricingManager(this);
    this.llmLogsManager = new LlmLogsManager(this);
    this.ordersManager = new OrdersManager(this);
    this.memberManager = new MemberManager(this);

    // 加载 dashboard
    this.dashboardManager.loadDashboard();
  }

  // 通用API调用方法，自动处理401错误
  async apiCall(url, options = {}) {
    const defaultOptions = {
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
        ...options.headers,
      },
    };

    const response = await fetch(url, {
      ...options,
      headers: defaultOptions.headers,
    });

    // 如果遇到401错误，自动跳转到登录页
    if (response.status === 401) {
      this.handleUnauthorized();
      throw new Error("Unauthorized");
    }
    return response.json();
  }

  showAlert(message, type = "info") {
    const kindConfig = new Map([
      ["info", { bgColor: "bg-blue-500", icon: "fa-info-circle" }],
      ["success", { bgColor: "bg-green-500", icon: "fa-check-circle" }],
      ["error", { bgColor: "bg-red-500", icon: "fa-exclamation-circle" }],
      [
        "warning",
        { bgColor: "bg-yellow-500", icon: "fa-exclamation-triangle" },
      ],
    ]);
    const config = kindConfig.get(type) || kindConfig.get("info");
    const eleObj = document.createElement("div");
    eleObj.className = `fixed top-5 right-4 p-4 rounded-lg shadow-lg z-50 ${config.bgColor} text-white`;
    eleObj.innerHTML = `
      <div class="flex items-center">
        <i class="fas ${config.icon} mr-2"></i>
        <span>${message}</span>
        <button class="ml-4 text-white hover:text-gray-200" onclick="this.parentElement.parentElement.remove()">
          <i class="fas fa-times"></i>
        </button>
      </div>
    `;
    document.body.appendChild(eleObj);
    setTimeout(() => {
      eleObj.parentElement && eleObj.remove();
    }, 5000);
  }

  // 处理401未授权错误
  handleUnauthorized() {
    this.token = null;
    this.apiKey = null;
    localStorage.removeItem("token");
    localStorage.removeItem("apiKey");
    this.showSignInPage();

    // 显示提示信息
    const errorDiv = document.getElementById("signInError");
    errorDiv.textContent = "登录已过期，请重新登录";
    errorDiv.classList.remove("hidden");
  }

  // HTML 转义辅助方法，防止 XSS 攻击
  escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }

  bindEvents() {
    // 登录表单
    document.getElementById("signInForm").addEventListener("submit", (e) => {
      e.preventDefault();
      this.signIn();
    });

    // 登出按钮
    document.getElementById("signOutBtn").addEventListener("click", () => {
      this.logout();
    });

    // 侧边栏切换
    document.getElementById("sidebarToggle").addEventListener("click", () => {
      this.toggleSidebar();
    });

    // 遮罩层点击
    document.getElementById("overlay").addEventListener("click", () => {
      this.closeSidebar();
    });

    // 导航菜单
    document.querySelectorAll(".nav-item").forEach((item) => {
      item.addEventListener("click", (e) => {
        e.preventDefault();
        const page = item.getAttribute("data-page");
        this.showPage(page);
      });
    });

    // 聊天测试
    document.getElementById("sendMessage").addEventListener("click", () => {
      this.sendChatMessage();
    });

    // 回车发送消息
    document
      .getElementById("messageInput")
      .addEventListener("keypress", (e) => {
        if (e.key === "Enter" && e.ctrlKey) {
          this.sendChatMessage();
        }
      });

    // 模型筛选按钮
    document.addEventListener("click", (e) => {
      if (e.target.classList.contains("provider-filter-btn")) {
        this.filterModelsByProvider(e.target.getAttribute("data-provider"));

        // 更新按钮样式
        document.querySelectorAll(".provider-filter-btn").forEach((btn) => {
          btn.classList.remove("bg-blue-500", "text-white");
          btn.classList.add("bg-gray-200", "text-gray-700");
        });

        e.target.classList.remove("bg-gray-200", "text-gray-700");
        e.target.classList.add("bg-blue-500", "text-white");
      }
    });

  }

  async loadUserInfo() {
    try {
      const resp = await this.apiCall("/api/admin/current");
      if (resp && resp.user) {
        this.apiKey = resp.user.api_key;
        localStorage.setItem("apiKey", this.apiKey);
      }
    } catch (error) {
      console.error("Failed to load user info:", error);
    }
  }

  async checkAuth() {
    if (this.token) {
      // 如果没有 API Key，尝试加载用户信息
      if (!this.apiKey) {
        await this.loadUserInfo();
      }
      this.showMainApp();
    } else {
      this.showSignInPage();
    }
  }

  async signIn() {
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
    const errorDiv = document.getElementById("signInError");

    try {
      const response = await fetch("/api/admin/signin", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      });

      const data = await response.json();

      if (response.ok) {
        this.token = data.token;
        localStorage.setItem("token", this.token);

        // 获取用户信息，包括 API Key
        await this.loadUserInfo();

        this.showMainApp();
        this.loadDashboard();
        errorDiv.classList.add("hidden");
      } else {
        errorDiv.textContent = data.error || "登录失败";
        errorDiv.classList.remove("hidden");
      }
    } catch (error) {
      errorDiv.textContent = "网络错误，请重试";
      errorDiv.classList.remove("hidden");
    }
  }

  async logout() {
    try {
      await fetch("/api/logout", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
        },
      });
    } catch (error) {
      console.error("Logout error:", error);
    }

    this.token = null;
    this.apiKey = null;
    localStorage.removeItem("token");
    localStorage.removeItem("apiKey");
    this.showSignInPage();
  }

  showSignInPage() {
    document.getElementById("signInPage").classList.remove("hidden");
    document.getElementById("mainApp").classList.add("hidden");
  }

  showMainApp() {
    document.getElementById("signInPage").classList.add("hidden");
    document.getElementById("mainApp").classList.remove("hidden");
  }

  toggleSidebar() {
    const sidebar = document.getElementById("sidebar");
    const overlay = document.getElementById("overlay");

    // 在小屏幕下切换显示/隐藏
    sidebar.classList.toggle("hidden");
    sidebar.classList.toggle("-translate-x-full");
    overlay.classList.toggle("hidden");
  }

  closeSidebar() {
    const sidebar = document.getElementById("sidebar");
    const overlay = document.getElementById("overlay");

    // 在小屏幕下隐藏侧边栏
    sidebar.classList.add("hidden");
    sidebar.classList.add("-translate-x-full");
    overlay.classList.add("hidden");
  }

  showPage(page) {
    // 隐藏所有页面
    document.querySelectorAll(".page-content").forEach((p) => {
      p.classList.add("hidden");
    });

    // 显示目标页面
    document.getElementById(`${page}Page`).classList.remove("hidden");

    // 更新导航状态
    document.querySelectorAll(".nav-item").forEach((item) => {
      item.classList.remove("bg-gray-700", "text-white");
      item.classList.add("text-gray-300");
    });

    document
      .querySelector(`[data-page="${page}"]`)
      .classList.add("bg-gray-700", "text-white");
    document
      .querySelector(`[data-page="${page}"]`)
      .classList.remove("text-gray-300");

    this.currentPage = page;

    // 加载页面数据
    switch (page) {
      case "dashboard":
        this.dashboardManager.loadDashboard();
        break;
      case "chat":
        this.testChatManager.loadModels();
        break;
      case "logs":
        this.llmLogsManager.loadLogsPage();
        break;
      case "models":
        this.loadModelsPage();
        break;
      case "users":
        this.memberManager.loadUsersPage();
        break;
      case "orders":
        this.ordersManager.loadOrdersPage();
        break;
      case "pricing-plans":
        this.pricingManager.loadPricingPlans();
        break;
      default:
        console.warn("Unknown page:", page);
    }

    // 在移动端关闭侧边栏
    if (window.innerWidth < 1024) {
      this.closeSidebar();
    }
  }

  async loadModelsPage() {
    try {
      const data = await this.apiCall("/api/admin/models");
      this.allModels = data.data || []; // 存储所有模型数据
      this.updateModelsTable(this.allModels);
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("Failed to load models page:", error);
      }
    }
  }

  updateModelsTable(models) {
    const tableDiv = document.getElementById("modelsTable");

    if (models.length === 0) {
      tableDiv.innerHTML =
        '<p class="text-gray-500 text-center">暂无可用模型</p>';
      return;
    }

    const table = document.createElement("div");
    table.className = "overflow-x-auto";
    table.innerHTML = `
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">模型 ID</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">提供商</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    ${models
                      .map(
                        (model) => `
                        <tr>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">${model.id}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${model.name}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                                    ${model.provider}
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                                    可用
                                </span>
                            </td>
                        </tr>
                    `
                      )
                      .join("")}
                </tbody>
            </table>
        `;

    tableDiv.innerHTML = "";
    tableDiv.appendChild(table);
  }

  // 按提供商筛选模型
  filterModelsByProvider(provider) {
    let filteredModels;

    if (provider === "all") {
      filteredModels = this.allModels;
    } else {
      filteredModels = this.allModels.filter(
        (model) => model.provider === provider
      );
    }

    this.updateModelsTable(filteredModels);
  }

}

// 初始化应用
document.addEventListener("DOMContentLoaded", () => {
  window.app = new LLMMemberApp();
});
