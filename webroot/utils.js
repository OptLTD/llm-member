// 全局状态管理
const AppState = {
  user: null,
  token: localStorage.getItem("userToken"),
  isLoggedIn: false,
};

// 认证管理
const Auth = {
  // 登录
  async signin(username, password) {
    try {
      const data = await Utils.apiRequest("/api/signin", {
        body: JSON.stringify({ username, password }),
      });

      AppState.token = data.token;
      AppState.user = data.user;
      AppState.isLoggedIn = true;

      localStorage.setItem("userToken", data.token);
      localStorage.setItem("userData", JSON.stringify(data.user));

      this.updateUI();
      Utils.showNotification("登录成功！", "success");

      return data;
    } catch (error) {
      Utils.showNotification(error.message, "error");
      throw error;
    }
  },

  // 注册
  async signup(email, username, password) {
    try {
      const data = await Utils.apiRequest("/api/signup", {
        body: JSON.stringify({ email, username, password }),
      });

      Utils.showNotification("注册成功！请登录您的账户", "success");
      return data;
    } catch (error) {
      Utils.showNotification(error.message, "error");
      throw error;
    }
  },

  // 退出登录
  async signout() {
    AppState.token = null;
    AppState.user = null;
    AppState.isLoggedIn = false;

    localStorage.removeItem("userToken");
    localStorage.removeItem("userData");

    this.updateUI();
  },

  // 检查登录状态
  async checkAuth() {
    const token = localStorage.getItem("userToken");
    const userData = localStorage.getItem("userData");

    if (token && userData) {
      try {
        AppState.token = token;
        AppState.user = JSON.parse(userData);
        AppState.isLoggedIn = true;
        this.updateUI();
      } catch (error) {
        console.error("检查登录状态失败:", error);
        this.logout();
      }
    } else {
      // 如果没有token或用户数据，确保状态为未登录
      AppState.token = null;
      AppState.user = null;
      AppState.isLoggedIn = false;
      this.updateUI();
    }
  },

  // 更新UI状态
  updateUI() {
    const userMenu = document.getElementById("userMenu");
    const userName = document.getElementById("userName");
    const signinBtn = document.getElementById("signInBtn");
    const signupBtn = document.getElementById("signUpBtn");

    if (AppState.isLoggedIn && AppState.user) {
      if (signupBtn) signupBtn.style.display = "none";
      if (signinBtn) signinBtn.style.display = "none";
      if (userMenu) userMenu.classList.remove("hidden");
      if (userName) userName.textContent = AppState.user.username;
    } else {
      if (signupBtn) signupBtn.style.display = "block";
      if (signinBtn) signinBtn.style.display = "block";
      if (userMenu) userMenu.classList.add("hidden");
    }
  },
};

// API 基础配置
const API_BASE = "/api";

// 工具函数
const Utils = {
  // 显示通知
  showNotification(message, type = "info") {
    const kindConfig = new Map([
      ["info", { bgColor: "bg-blue-500", icon: "fa-info-circle" }],
      ["success", { bgColor: "bg-green-500", icon: "fa-check-circle" }],
      ["error", { bgColor: "bg-red-500", icon: "fa-exclamation-circle" }],
      ["warning", { bgColor: "bg-yellow-500", icon: "fa-exclamation-triangle" }],
    ]);
    const config = kindConfig.get(type) || kindConfig.get("info");
    const notification = document.createElement("div");
    notification.className = `fixed top-20 right-4 p-4 rounded-lg shadow-lg z-50 ${config.bgColor} text-white`;
    notification.innerHTML = `
        <div class="flex items-center">
          <i class="fas ${config.icon} mr-2"></i>
          <span>${message}</span>
          <button class="ml-4 text-white hover:text-gray-200" onclick="this.parentElement.parentElement.remove()">
            <i class="fas fa-times"></i>
          </button>
        </div>
    `;
    document.body.appendChild(notification);

    // 自动移除通知
    setTimeout(() => {
      if (notification.parentElement) {
        notification.remove();
      }
    }, 5000);
  },

  // API 请求封装
  async apiRequest(url, options = {}) {
    const token = localStorage.getItem("userToken");
    const defaultOptions = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token && { Authorization: `Bearer ${token}` }),
      },
    };

    const response = await fetch(url, {
      ...defaultOptions,
      ...options,
      headers: {
        ...defaultOptions.headers,
        ...options.headers,
      },
    });

    const data = await response.json();

    if (!response.ok) {
      // 如果是401未授权错误，自动跳转到登录页面
      if (response.status === 401) {
        Auth.signout();
        const path = location.pathname
        localStorage.setItem("prevPage", path);
        window.location.href = "/signin";
        throw new Error("认证已过期，请重新登录");
      }
      throw new Error(data.error || "请求失败");
    }

    return data;
  },

  // 格式化数字
  formatNumber(num) {
    return new Intl.NumberFormat("zh-CN").format(num);
  },

  // 格式化日期
  formatDate(dateString) {
    return new Date(dateString).toLocaleString("zh-CN");
  },

  // 检查密码强度
  checkPasswordStrength(password) {
    if (!password) {
      return { strength: 0, feedback: '请输入密码' };
    }

    let score = 0;
    const feedback = [];

    // 长度检查
    if (password.length >= 8) {
      score += 1;
    } else {
      feedback.push('密码长度至少8位');
    }

    // 包含小写字母
    if (/[a-z]/.test(password)) {
      score += 1;
    } else {
      feedback.push('包含小写字母');
    }

    // 包含大写字母
    if (/[A-Z]/.test(password)) {
      score += 1;
    } else {
      feedback.push('包含大写字母');
    }

    // 包含数字
    if (/\d/.test(password)) {
      score += 1;
    } else {
      feedback.push('包含数字');
    }

    // 包含特殊字符
    if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      score += 1;
    } else {
      feedback.push('包含特殊字符');
    }

    // 根据得分返回强度等级和反馈
    const strengthLevels = [
      { level: 0, text: '非常弱', color: 'red' },
      { level: 1, text: '弱', color: 'orange' },
      { level: 2, text: '一般', color: 'yellow' },
      { level: 3, text: '良好', color: 'blue' },
      { level: 4, text: '强', color: 'green' },
      { level: 5, text: '非常强', color: 'green' }
    ];

    const strengthInfo = strengthLevels[score];
    
    return {
      strength: score,
      level: strengthInfo.level,
      text: strengthInfo.text,
      color: strengthInfo.color,
      feedback: feedback.length > 0 ? `建议: ${feedback.join('、')}` : '密码强度良好'
    };
  },
};

// 事件监听器管理
const EventListeners = {
  // 首页事件监听器
  setupHomeEventListeners() {
    // 开始使用按钮
    const getStartedBtn = document.getElementById("getStartedBtn");
    if (getStartedBtn) {
      getStartedBtn.addEventListener("click", () => {
        if (AppState.isLoggedIn) {
          window.location.hash = "#profile";
        } else {
          window.location.href = "/signup";
        }
      });
    }

    // CTA区域的注册按钮
    const ctaStartBtn = document.getElementById("ctaStartBtn");
    if (ctaStartBtn) {
      ctaStartBtn.addEventListener("click", () => {
        if (AppState.isLoggedIn) {
          window.location.hash = "/profile";
        } else {
          window.location.href = "/signup";
        }
      });
    }

    // 导航链接平滑滚动
    document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
      anchor.addEventListener("click", function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute("href"));
        if (target) {
          target.scrollIntoView({
            behavior: "smooth",
            block: "start",
          });
        }
      });
    });
  },

  // 个人中心事件监听器
  setupProfileEventListeners() {
    // 侧边栏头像点击事件
    const sidebarUserName = document.getElementById("sidebarUserName");
    if (sidebarUserName) {
      sidebarUserName.addEventListener("click", () => {
        const profileModal = document.getElementById("profileModal");
        if (profileModal) {
          profileModal.classList.remove("hidden");
        }
      });
    }

    const sidebarAvatar = document.getElementById("sidebarAvatar");
    if (sidebarAvatar) {
      sidebarAvatar.addEventListener("click", () => {
        const profileModal = document.getElementById("profileModal");
        if (profileModal) {
          profileModal.classList.remove("hidden");
        }
      });
    }

    // 关闭模态框
    const closeProfileModal = document.getElementById("closeProfileModal");
    if (closeProfileModal) {
      closeProfileModal.addEventListener("click", () => {
        const profileModal = document.getElementById("profileModal");
        if (profileModal) {
          profileModal.classList.add("hidden");
        }
      });
    }

    const cancelProfileEdit = document.getElementById("cancelProfileEdit");
    if (cancelProfileEdit) {
      cancelProfileEdit.addEventListener("click", () => {
        const profileModal = document.getElementById("profileModal");
        if (profileModal) {
          profileModal.classList.add("hidden");
        }
      });
    }

    // 点击模态框外部关闭
    const profileModal = document.getElementById("profileModal");
    if (profileModal) {
      profileModal.addEventListener("click", (e) => {
        if (e.target === profileModal) {
          profileModal.classList.add("hidden");
        }
      });
    }

    // 保存个人资料
    const saveProfileBtn = document.getElementById("saveProfileBtn");
    if (saveProfileBtn) {
      saveProfileBtn.addEventListener("click", async () => {
        const email = document.getElementById("modalProfileEmail").value;
        if (window.DataManager) {
          await window.DataManager.updateUserProfile({ email });
        }
        const profileModal = document.getElementById("profileModal");
        if (profileModal) {
          profileModal.classList.add("hidden");
        }
      });
    }

    // 个人资料表单提交
    const profileForm = document.getElementById("profileForm");
    if (profileForm) {
      profileForm.addEventListener("submit", async (e) => {
        e.preventDefault();

        const email = document.getElementById("profileEmail").value;

        if (window.DataManager) {
          await window.DataManager.updateUserProfile({ email });
        }
      });
    }

    // API密钥相关按钮
    const toggleApiKeyBtn = document.getElementById("toggleApiKeyBtn");
    if (toggleApiKeyBtn) {
      toggleApiKeyBtn.addEventListener("click", () => {
        const apiKeyDisplay = document.getElementById("apiKeyDisplay");
        const toggleBtn = document.getElementById("toggleApiKeyBtn");

        if (apiKeyDisplay && toggleBtn) {
          if (apiKeyDisplay.type === "password") {
            apiKeyDisplay.type = "text";
            toggleBtn.innerHTML = '<i class="fas fa-eye-slash"></i>';
          } else {
            apiKeyDisplay.type = "password";
            toggleBtn.innerHTML = '<i class="fas fa-eye"></i>';
          }
        }
      });
    }

    const copyApiKeyBtn = document.getElementById("copyApiKeyBtn");
    if (copyApiKeyBtn) {
      copyApiKeyBtn.addEventListener("click", () => {
        const apiKey = document.getElementById("apiKeyDisplay").value;
        Utils.copyToClipboard(apiKey);
      });
    }

    const regenerateApiKeyBtn = document.getElementById("regenerateApiKeyBtn");
    if (regenerateApiKeyBtn) {
      regenerateApiKeyBtn.addEventListener("click", () => {
        if (window.DataManager) {
          window.DataManager.regenerateAPIKey();
        }
      });
    }

    // 跳转到支付页面
    const goToPaymentBtn = document.getElementById("goToPaymentBtn");
    if (goToPaymentBtn) {
      goToPaymentBtn.addEventListener("click", () => {
        window.location.href = "/payment";
      });
    }
  },

  // 价格页面事件监听器
  setupPricingEventListeners() {
    // 价格页面特有的事件监听器可以在这里添加
    // 目前价格页面没有特有的事件监听器
  },

  // 注册页面事件监听器
  setupSignupEventListeners() {
    // 注册页面保持原有的事件监听器，不做调整
    // 页面自身的事件监听器继续在页面内部处理
  },

  // 登录页面事件监听器
  setupSigninEventListeners() {
    // 登录页面保持原有的事件监听器，不做调整
    // 页面自身的事件监听器继续在页面内部处理
  },

  // 通用事件监听器设置
  setupEventListeners() {
    // 根据当前页面路径决定设置哪些事件监听器
    const path = window.location.pathname;
    if (path === "/" || path === "/index.html") {
      this.setupHomeEventListeners();
    } else if (path === "/profile" || path === "/profile.html") {
      this.setupProfileEventListeners();
    } else if (path === "/pricing" || path === "/pricing.html") {
      this.setupPricingEventListeners();
    } else if (path === "/signup" || path === "/signup.html") {
      this.setupSignupEventListeners();
    } else if (path === "/signin" || path === "/signin.html") {
      this.setupSigninEventListeners();
    }

    // 设置通用的事件监听器（所有页面都需要的）
    this.setupCommonEventListeners();
  },

  // 通用事件监听器（所有页面都需要的）
  setupCommonEventListeners() {
    const userMenuBtn = document.getElementById("userMenuBtn");
    const userDropdown = document.getElementById("userDropdown");
    if (userMenuBtn && userDropdown) {
      userMenuBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        userDropdown.classList.toggle("hidden");
      });

      // 点击外部关闭下拉菜单（所有页面通用）
      document.addEventListener("click", (e) => {
        if (
          !userMenuBtn.contains(e.target) &&
          !userDropdown.contains(e.target)
        ) {
          userDropdown.classList.add("hidden");
        }
      });
    }

    // 通用的登录按钮处理
    const signInBtn = document.getElementById("signInBtn");
    if (signInBtn) {
      signInBtn.addEventListener("click", () => {
        window.location.href = "/signin";
      });
    }

    // 通用的登出按钮处理
    const signOutBtn = document.getElementById("signOutBtn");
    if (signOutBtn) {
      signOutBtn.addEventListener("click", () => {
        if (window.App && window.App.Auth) {
          window.App.Auth.signout();
          window.location.href = "/";
        } else {
          localStorage.removeItem("userToken");
          localStorage.removeItem("userData");
          window.location.href = "/";
        }
      });
    }
  },
};

window.App = {
  Auth,
  Utils,
  AppState,
  EventListeners,
};