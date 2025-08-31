// 订单管理模块
class OrdersManager {
  constructor(app) {
    this.app = app;
    this.currentOrderPage = 1;
    this.ordersPerPage = 20;
    this.totalOrders = 0;
    this.bindEvents();
  }

  bindEvents() {
    // 订单管理相关事件
    const searchOrdersBtn = document.getElementById("searchOrdersBtn");
    const orderStatusFilter = document.getElementById("orderStatusFilter");
    const paymentMethodFilter = document.getElementById("paymentMethodFilter");
    const userSearchInput = document.getElementById("userSearchInput");
    const prevPageBtn = document.getElementById("prevPageBtn");
    const nextPageBtn = document.getElementById("nextPageBtn");

    if (searchOrdersBtn) {
      searchOrdersBtn.addEventListener("click", () => {
        this.currentOrderPage = 1;
        this.loadOrdersPage();
      });
    }

    if (orderStatusFilter) {
      orderStatusFilter.addEventListener("change", () => {
        this.currentOrderPage = 1;
        this.loadOrdersPage();
      });
    }

    if (paymentMethodFilter) {
      paymentMethodFilter.addEventListener("change", () => {
        this.currentOrderPage = 1;
        this.loadOrdersPage();
      });
    }

    if (userSearchInput) {
      userSearchInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.currentOrderPage = 1;
          this.loadOrdersPage();
        }
      });
    }

    if (prevPageBtn) {
      prevPageBtn.addEventListener("click", () => {
        if (this.currentOrderPage > 1) {
          this.currentOrderPage--;
          this.loadOrdersPage();
        }
      });
    }

    if (nextPageBtn) {
      nextPageBtn.addEventListener("click", () => {
        const totalPages = Math.ceil(this.totalOrders / this.ordersPerPage);
        if (this.currentOrderPage < totalPages) {
          this.currentOrderPage++;
          this.loadOrdersPage();
        }
      });
    }
  }

  changePage(page) {
    this.currentOrderPage = page;
    this.loadOrdersPage();
  }

  async loadOrdersPage() {
    try {
      const userSearch = document.getElementById("userSearchInput").value;
      const statusFilter = document.getElementById("orderStatusFilter").value;
      const methodFilter = document.getElementById("paymentMethodFilter").value;
      
      const params = {
        page: this.currentOrderPage,
        size: this.ordersPerPage,
        query: {},
      };
      
      if (userSearch) params.query.search = userSearch;
      if (statusFilter) params.query.status = statusFilter;
      if (methodFilter) params.query.method = methodFilter;
      
      const resp = await this.app.apiCall(`/api/admin/orders`, {
        method: "POST", 
        body: JSON.stringify(params)
      });
      
      if (resp.data) {
        this.totalOrders = resp.total;
        this.updateOrdersTable(resp.data);
        this.updateOrdersPagination();
      } else {
        this.app.showAlert(resp.message || "加载订单失败", "error");
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("加载订单失败:", error);
        this.app.showAlert("加载订单失败", "error");
      }
    }
  }

  updateOrdersTable(orders) {
    const tableContainer = document.getElementById("ordersTable");
    
    if (!orders || orders.length === 0) {
      tableContainer.innerHTML = `
        <div class="text-center py-8">
          <i class="fas fa-shopping-cart text-4xl text-gray-300 mb-4"></i>
          <p class="text-gray-500">暂无订单数据</p>
        </div>
      `;
      return;
    }

    const table = `
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">订单ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">方案</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">支付方式</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            ${orders.map(order => `
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  ${this.app.escapeHtml(order.id || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">${this.app.escapeHtml(order.username || '')}</div>
                  <div class="text-sm text-gray-500">${this.app.escapeHtml(order.email || '')}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.app.escapeHtml(order.planName || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ¥${(order.amount || 0).toFixed(2)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  ${this.getPaymentMethodBadge(order.paymentMethod)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  ${this.getOrderStatusBadge(order.status)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  ${this.formatDateTime(order.createdAt)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button onclick="app.ordersManager.viewOrderDetails('${order.id}')" 
                    class="text-blue-600 hover:text-blue-900 mr-3">
                    <i class="fas fa-eye"></i> 查看
                  </button>
                  ${order.status === 'pending' ? `
                    <button onclick="app.ordersManager.cancelOrder('${order.id}')" 
                      class="text-red-600 hover:text-red-900">
                      <i class="fas fa-times"></i> 取消
                    </button>
                  ` : ''}
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
    
    tableContainer.innerHTML = table;
  }

  getPaymentMethodBadge(method) {
    const badges = {
      'alipay': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">支付宝</span>',
      'wechat': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">微信支付</span>',
      'paypal': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">PayPal</span>',
      'stripe': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-violet-100 text-violet-800">Stripe</span>',
      'creem': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">Creem</span>'
    };
    return badges[method] || `<span class="text-gray-500">${this.app.escapeHtml(method || '')}</span>`;
  }

  getOrderStatusBadge(status) {
    const badges = {
      'pending': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">待支付</span>',
      'paid': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">已支付</span>',
      'cancelled': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">已取消</span>',
      'expired': '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">已过期</span>'
    };
    return badges[status] || `<span class="text-gray-500">${this.app.escapeHtml(status || '')}</span>`;
  }

  updateOrdersPagination() {
    const totalPages = Math.ceil(this.totalOrders / this.ordersPerPage);
    const startItem = (this.currentOrderPage - 1) * this.ordersPerPage + 1;
    const endItem = Math.min(this.currentOrderPage * this.ordersPerPage, this.totalOrders);
    
    // 更新分页信息
    const currentPageStart = document.getElementById("currentPageStart");
    const currentPageEnd = document.getElementById("currentPageEnd");
    const totalOrdersCount = document.getElementById("totalOrdersCount");
    
    if (currentPageStart) currentPageStart.textContent = this.totalOrders > 0 ? startItem : 0;
    if (currentPageEnd) currentPageEnd.textContent = this.totalOrders > 0 ? endItem : 0;
    if (totalOrdersCount) totalOrdersCount.textContent = this.totalOrders;
    
    // 更新按钮状态
    const prevPageBtn = document.getElementById("prevPageBtn");
    const nextPageBtn = document.getElementById("nextPageBtn");
    
    if (prevPageBtn) {
      prevPageBtn.disabled = this.currentOrderPage <= 1;
    }
    
    if (nextPageBtn) {
      nextPageBtn.disabled = this.currentOrderPage >= totalPages;
    }
    
    // 生成页码按钮
    this.generatePageNumbers(totalPages);
  }

  generatePageNumbers(totalPages) {
    const pageNumbersContainer = document.getElementById("pageNumbers");
    if (!pageNumbersContainer) return;
    
    pageNumbersContainer.innerHTML = '';
    
    if (totalPages <= 1) return;
    
    const maxVisiblePages = 5;
    let startPage = Math.max(1, this.currentOrderPage - Math.floor(maxVisiblePages / 2));
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
    
    if (endPage - startPage + 1 < maxVisiblePages) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }
    
    for (let i = startPage; i <= endPage; i++) {
      const pageBtn = document.createElement('button');
      pageBtn.className = `px-3 py-1 border rounded ${
        i === this.currentOrderPage 
          ? 'bg-blue-500 text-white border-blue-500' 
          : 'border-gray-300 hover:bg-gray-50'
      }`;
      pageBtn.textContent = i;
      pageBtn.addEventListener('click', () => {
        this.currentOrderPage = i;
        this.loadOrdersPage();
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

  async viewOrderDetails(orderId) {
    try {
      const response = await this.app.apiCall(`/api/admin/orders/${orderId}`);
      if (response.success) {
        // 这里可以显示订单详情弹窗
        this.app.showAlert(`订单详情: ${JSON.stringify(response.data, null, 2)}`, "info");
      } else {
        this.app.showAlert(response.message || "获取订单详情失败", "error");
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("获取订单详情失败:", error);
        this.app.showAlert("获取订单详情失败", "error");
      }
    }
  }

  async cancelOrder(orderId) {
    if (!confirm('确定要取消这个订单吗？')) {
      return;
    }
    
    try {
      const response = await this.app.apiCall(`/api/admin/orders/${orderId}/cancel`, {
        method: 'POST'
      });
      
      if (response.success) {
        this.app.showAlert("订单已取消", "success");
        this.loadOrdersPage();
      } else {
        this.app.showAlert(response.message || "取消订单失败", "error");
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("取消订单失败:", error);
        this.app.showAlert("取消订单失败", "error");
      }
    }
  }
}