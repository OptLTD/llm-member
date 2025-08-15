// 用户管理模块
class MemberManager {
  constructor(app) {
    this.app = app;
    this.currentUserPage = 1;
    this.usersPerPage = 20;
    this.totalUsers = 0;
    this.bindEvents();
  }

  bindEvents() {
    // 用户注册表单
    document
      .getElementById("createUserForm")
      .addEventListener("submit", (e) => {
        e.preventDefault();
        this.createUser();
      });

    // 绑定添加用户弹窗
    const addUserBtn = document.getElementById("addUserBtn");
    const addUserModal = document.getElementById("addUserModal");
    const closeModalBtn = document.getElementById("closeModalBtn");
    const cancelBtn = document.getElementById("cancelBtn");

    if (addUserBtn && addUserModal) {
      addUserBtn.addEventListener("click", () => {
        this.showAddUserModal();
      });
    }

    if (closeModalBtn && addUserModal) {
      closeModalBtn.addEventListener("click", () => {
        this.hideAddUserModal();
      });
    }

    if (cancelBtn && addUserModal) {
      cancelBtn.addEventListener("click", () => {
        this.hideAddUserModal();
      });
    }

    // 点击弹窗外部关闭
    if (addUserModal) {
      addUserModal.addEventListener("click", (e) => {
        if (e.target === addUserModal) {
          this.hideAddUserModal();
        }
      });
    }

    // 筛选和搜索相关事件
    const searchUsersBtn = document.getElementById("searchUsersBtn");
    const userStatusFilter = document.getElementById("userStatusFilter");
    const userRoleFilter = document.getElementById("userRoleFilter");
    const userNameSearchInput = document.getElementById("userNameSearchInput");
    const prevUserPageBtn = document.getElementById("prevUserPageBtn");
    const nextUserPageBtn = document.getElementById("nextUserPageBtn");

    if (searchUsersBtn) {
      searchUsersBtn.addEventListener("click", () => {
        this.currentUserPage = 1;
        this.loadUsersPage();
      });
    }

    if (userStatusFilter) {
      userStatusFilter.addEventListener("change", () => {
        this.currentUserPage = 1;
        this.loadUsersPage();
      });
    }

    if (userRoleFilter) {
      userRoleFilter.addEventListener("change", () => {
        this.currentUserPage = 1;
        this.loadUsersPage();
      });
    }

    if (userNameSearchInput) {
      userNameSearchInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.currentUserPage = 1;
          this.loadUsersPage();
        }
      });
    }

    if (prevUserPageBtn) {
      prevUserPageBtn.addEventListener("click", () => {
        if (this.currentUserPage > 1) {
          this.currentUserPage--;
          this.loadUsersPage();
        }
      });
    }

    if (nextUserPageBtn) {
      nextUserPageBtn.addEventListener("click", () => {
        const totalPages = Math.ceil(this.totalUsers / this.usersPerPage);
        if (this.currentUserPage < totalPages) {
          this.currentUserPage++;
          this.loadUsersPage();
        }
      });
    }
  }

  showAddUserModal() {
    const modal = document.getElementById("addUserModal");
    modal.classList.remove("hidden");

    // 清空表单
    document.getElementById("createUserForm").reset();
    document.getElementById("newRequestLimit").value = "100";

    // 隐藏错误和成功消息
    document.getElementById("createUserError").classList.add("hidden");
    document.getElementById("createUserSuccess").classList.add("hidden");
  }

  hideAddUserModal() {
    const modal = document.getElementById("addUserModal");
    modal.classList.add("hidden");
  }

  async createUser() {
    const username = document.getElementById("newUsername").value;
    const email = document.getElementById("newEmail").value;
    const password = document.getElementById("newPassword").value;
    const requestLimit = parseInt(
      document.getElementById("newRequestLimit").value
    );
    const isAdmin = document.getElementById("newIsAdmin").checked;

    const errorDiv = document.getElementById("createUserError");
    const successDiv = document.getElementById("createUserSuccess");

    try {
      const data = await this.app.apiCall("/api/admin/users/create", {
        method: "POST",
        body: JSON.stringify({
          username, email, password,
          request_limit: requestLimit,
          curr_role: isAdmin ? 'admin': 'user',
        }),
      });
      successDiv.textContent = `用户创建成功！API Key: ${data.api_key}`;
      successDiv.classList.remove("hidden");
      errorDiv.classList.add("hidden");

      // 重新加载用户列表
      this.loadUsersPage();

      // 延迟关闭弹窗
      setTimeout(() => {
        this.hideAddUserModal();
      }, 2000);
    } catch (error) {
      if (error.message !== "Unauthorized") {
        errorDiv.textContent = error.message || "网络错误，请重试";
        errorDiv.classList.remove("hidden");
        successDiv.classList.add("hidden");
      }
    }
  }

  changePage(page) {
    this.currentUserPage = page;
    this.loadUsersPage();
  }

  async loadUsersPage() {
    try {
      const userNameSearch = document.getElementById("userNameSearchInput").value;
      const statusFilter = document.getElementById("userStatusFilter").value;
      const roleFilter = document.getElementById("userRoleFilter").value;
      
      const params = {
        page: this.currentUserPage,
        size: this.usersPerPage,
        query: {},
      };
      
      if (userNameSearch) params.query.search = userNameSearch;
      if (statusFilter) params.query.status = statusFilter;
      if (roleFilter) params.query.role = roleFilter;
      
      const resp = await this.app.apiCall(`/api/admin/users`, {
        method: "POST", body: JSON.stringify(params)
      });
      
      if (resp.data) {
        this.totalUsers = resp.total;
        this.updateUsersTable(resp.data);
        this.updateUsersPagination();
      } else {
        this.app.showAlert(resp.message || "加载用户失败", "error");
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("加载用户失败:", error);
        this.app.showAlert("加载用户失败", "error");
      }
    }
  }

  updateUsersTable(users) {
    const tableContainer = document.getElementById("usersTable");
    
    if (!users || users.length === 0) {
      tableContainer.innerHTML = `
        <div class="text-center py-8">
          <i class="fas fa-users text-4xl text-gray-300 mb-4"></i>
          <p class="text-gray-500">暂无用户数据</p>
        </div>
      `;
      return;
    }

    const table = `
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">用户名</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">邮箱</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">API Key</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">权限</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">请求统计</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            ${users.map(user => `
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  ${this.app.escapeHtml(user.username || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.app.escapeHtml(user.email || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-600">
                  <span class="truncate block w-32" title="${this.app.escapeHtml(user.api_key || '')}">
                    ${user.api_key ? user.api_key.substring(0, 20) + '...' : ''}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                    user.is_active
                      ? "bg-green-100 text-green-800"
                      : "bg-red-100 text-red-800"
                  }">
                    ${user.is_active ? "活跃" : "禁用"}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                    user.curr_role == 'admin'
                      ? "bg-purple-100 text-purple-800"
                      : "bg-gray-100 text-gray-800"
                  }">
                    ${user.curr_role == 'admin' ? "管理员" : "普通用户"}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${user.total_requests || 0} / ${user.request_limit || 0}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  ${this.formatDateTime(user.created_at)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button onclick="app.memberManager.regenerateAPIKey(${user.id})" 
                    class="text-blue-600 hover:text-blue-900 mr-3">
                    <i class="fas fa-key"></i> 重新生成Key
                  </button>
                  <button onclick="app.memberManager.toggleUserStatus(${user.id}, ${!user.is_active})" 
                    class="${
                      user.is_active
                        ? "text-red-600 hover:text-red-900"
                        : "text-green-600 hover:text-green-900"
                    }">
                    <i class="fas fa-${user.is_active ? 'ban' : 'check'}"></i> ${user.is_active ? "禁用" : "启用"}
                  </button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
    
    tableContainer.innerHTML = table;
  }

  updateUsersPagination() {
    const totalPages = Math.ceil(this.totalUsers / this.usersPerPage);
    const startItem = (this.currentUserPage - 1) * this.usersPerPage + 1;
    const endItem = Math.min(this.currentUserPage * this.usersPerPage, this.totalUsers);
    
    // 更新分页信息
    const currentUserPageStart = document.getElementById("currentUserPageStart");
    const currentUserPageEnd = document.getElementById("currentUserPageEnd");
    const totalUsersCount = document.getElementById("totalUsersCount");
    
    if (currentUserPageStart) currentUserPageStart.textContent = this.totalUsers > 0 ? startItem : 0;
    if (currentUserPageEnd) currentUserPageEnd.textContent = this.totalUsers > 0 ? endItem : 0;
    if (totalUsersCount) totalUsersCount.textContent = this.totalUsers;
    
    // 更新按钮状态
    const prevUserPageBtn = document.getElementById("prevUserPageBtn");
    const nextUserPageBtn = document.getElementById("nextUserPageBtn");
    
    if (prevUserPageBtn) {
      prevUserPageBtn.disabled = this.currentUserPage <= 1;
    }
    
    if (nextUserPageBtn) {
      nextUserPageBtn.disabled = this.currentUserPage >= totalPages;
    }
    
    // 生成页码按钮
    this.generateUserPageNumbers(totalPages);
  }

  generateUserPageNumbers(totalPages) {
    const pageNumbersContainer = document.getElementById("userPageNumbers");
    if (!pageNumbersContainer) return;
    
    pageNumbersContainer.innerHTML = '';
    
    if (totalPages <= 1) return;
    
    const maxVisiblePages = 5;
    let startPage = Math.max(1, this.currentUserPage - Math.floor(maxVisiblePages / 2));
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
    
    if (endPage - startPage + 1 < maxVisiblePages) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }
    
    for (let i = startPage; i <= endPage; i++) {
      const pageBtn = document.createElement('button');
      pageBtn.className = `px-3 py-1 border rounded ${
        i === this.currentUserPage 
          ? 'bg-blue-500 text-white border-blue-500' 
          : 'border-gray-300 hover:bg-gray-50'
      }`;
      pageBtn.textContent = i;
      pageBtn.addEventListener('click', () => {
        this.currentUserPage = i;
        this.loadUsersPage();
      });
      pageNumbersContainer.appendChild(pageBtn);
    }
  }

  formatDateTime(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  async regenerateAPIKey(userId) {
    if (!confirm("确定要重新生成此用户的 API Key 吗？原 Key 将失效。")) {
      return;
    }

    try {
      const url = `/api/admin/users/${userId}/generate`;
      const data = await this.app.apiCall(url, { method: "POST" });
      this.app.showAlert(`新的 API Key: ${data.api_key}`, "success");
      this.loadUsersPage();
    } catch (error) {
      if (error.message !== "Unauthorized") {
        this.app.showAlert("操作失败: " + (error.message || "网络错误，请重试"));
      }
    }
  }

  async toggleUserStatus(userId, enable) {
    const action = enable ? "启用" : "禁用";
    if (!confirm(`确定要${action}此用户吗？`)) {
      return;
    }

    try {
      await this.app.apiCall(`/api/admin/users/${userId}/toggle`, {
        method: "POST",
        body: JSON.stringify({ is_active: enable }),
      });

      this.app.showAlert(`用户${action}成功`, "success");
      this.loadUsersPage();
    } catch (error) {
      if (error.message !== "Unauthorized") {
        this.app.showAlert(
          `${action}失败: ` + (error.message || "网络错误，请重试")
        );
      }
    }
  }
}