// 日志管理模块
class LlmLogsManager {
  constructor(app) {
    this.app = app;
    this.currentLogPage = 1;
    this.logsPerPage = 20;
    this.totalLogs = 0;
    this.bindEvents();
  }

  bindEvents() {
    // 筛选和搜索相关事件
    const searchLogsBtn = document.getElementById("searchLogsBtn");
    const logStatusFilter = document.getElementById("logStatusFilter");
    const logModelFilter = document.getElementById("logModelFilter");
    const logProviderFilter = document.getElementById("logProviderFilter");
    const logIpSearchInput = document.getElementById("logIpSearchInput");
    const prevLogPageBtn = document.getElementById("prevLogPageBtn");
    const nextLogPageBtn = document.getElementById("nextLogPageBtn");

    if (searchLogsBtn) {
      searchLogsBtn.addEventListener("click", () => {
        this.currentLogPage = 1;
        this.loadLogsPage();
      });
    }

    if (logStatusFilter) {
      logStatusFilter.addEventListener("change", () => {
        this.currentLogPage = 1;
        this.loadLogsPage();
      });
    }

    if (logModelFilter) {
      logModelFilter.addEventListener("change", () => {
        this.currentLogPage = 1;
        this.loadLogsPage();
      });
    }

    if (logProviderFilter) {
      logProviderFilter.addEventListener("change", () => {
        this.currentLogPage = 1;
        this.loadLogsPage();
      });
    }

    if (logIpSearchInput) {
      logIpSearchInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.currentLogPage = 1;
          this.loadLogsPage();
        }
      });
    }

    if (prevLogPageBtn) {
      prevLogPageBtn.addEventListener("click", () => {
        if (this.currentLogPage > 1) {
          this.currentLogPage--;
          this.loadLogsPage();
        }
      });
    }

    if (nextLogPageBtn) {
      nextLogPageBtn.addEventListener("click", () => {
        const totalPages = Math.ceil(this.totalLogs / this.logsPerPage);
        if (this.currentLogPage < totalPages) {
          this.currentLogPage++;
          this.loadLogsPage();
        }
      });
    }
  }

  changePage(page) {
    this.currentLogPage = page;
    this.loadLogsPage();
  }

  async loadLogsPage() {
    try {
      const ipSearch = document.getElementById("logIpSearchInput").value;
      const statusFilter = document.getElementById("logStatusFilter").value;
      const modelFilter = document.getElementById("logModelFilter").value;
      const providerFilter = document.getElementById("logProviderFilter").value;
      
      const params = {
        page: this.currentLogPage,
        size: this.logsPerPage,
        query: {},
      };
      
      if (ipSearch) params.query.ip = ipSearch;
      if (statusFilter) params.query.status = statusFilter;
      if (modelFilter) params.query.model = modelFilter;
      if (providerFilter) params.query.provider = providerFilter;
      
      const resp = await this.app.apiCall(`/api/admin/logs`, {
        method: "POST", body: JSON.stringify(params)
      });
      
      if (resp.data) {
        this.totalLogs = resp.total;
        this.updateLogsTable(resp.data);
        this.updateLogsPagination();
        this.loadFilterOptions(resp.data);
      } else {
        this.app.showAlert(resp.message || "加载日志失败", "error");
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("加载日志失败:", error);
        this.app.showAlert("加载日志失败", "error");
      }
    }
  }

  loadFilterOptions(logs) {
    // 提取唯一的模型和提供商选项
    const models = [...new Set(logs.map(log => log.model).filter(Boolean))];
    const providers = [...new Set(logs.map(log => log.provider).filter(Boolean))];
    
    // 更新模型筛选器
    const modelFilter = document.getElementById("logModelFilter");
    if (modelFilter) {
      const currentValue = modelFilter.value;
      modelFilter.innerHTML = '<option value="">全部模型</option>';
      models.forEach(model => {
        const option = document.createElement('option');
        option.value = model;
        option.textContent = model;
        if (model === currentValue) option.selected = true;
        modelFilter.appendChild(option);
      });
    }
    
    // 更新提供商筛选器
    const providerFilter = document.getElementById("logProviderFilter");
    if (providerFilter) {
      const currentValue = providerFilter.value;
      providerFilter.innerHTML = '<option value="">全部提供商</option>';
      providers.forEach(provider => {
        const option = document.createElement('option');
        option.value = provider;
        option.textContent = provider;
        if (provider === currentValue) option.selected = true;
        providerFilter.appendChild(option);
      });
    }
  }

  updateLogsTable(logs) {
    const tableContainer = document.getElementById("logsTable");
    
    if (!logs || logs.length === 0) {
      tableContainer.innerHTML = `
        <div class="text-center py-8">
          <i class="fas fa-file-alt text-4xl text-gray-300 mb-4"></i>
          <p class="text-gray-500">暂无日志数据</p>
        </div>
      `;
      return;
    }

    const table = `
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">模型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">提供商</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Token</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">耗时</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            ${logs.map(log => `
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.formatDateTime(log.created_at)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.app.escapeHtml(log.model || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.app.escapeHtml(log.provider || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  ${this.getLogStatusBadge(log.status)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${log.tokens_used || 0}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${log.duration || 0}ms
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${this.app.escapeHtml(log.client_ip || '')}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button onclick="app.logsManager.viewLogDetails(${log.id})" 
                    class="text-blue-600 hover:text-blue-900 mr-3">
                    <i class="fas fa-eye"></i> 详情
                  </button>
                  ${log.error_msg ? `
                    <button onclick="app.logsManager.viewLogError(${log.id})" 
                      class="text-red-600 hover:text-red-900">
                      <i class="fas fa-exclamation-triangle"></i> 错误
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

  getLogStatusBadge(status) {
    const statusConfig = {
      'success': { class: 'bg-green-100 text-green-800', text: '成功' },
      'error': { class: 'bg-red-100 text-red-800', text: '失败' },
      'pending': { class: 'bg-yellow-100 text-yellow-800', text: '处理中' }
    };
    
    const config = statusConfig[status] || { class: 'bg-gray-100 text-gray-800', text: status || '未知' };
    
    return `<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${config.class}">
      ${config.text}
    </span>`;
  }

  updateLogsPagination() {
    const totalPages = Math.ceil(this.totalLogs / this.logsPerPage);
    const startItem = (this.currentLogPage - 1) * this.logsPerPage + 1;
    const endItem = Math.min(this.currentLogPage * this.logsPerPage, this.totalLogs);
    
    // 更新分页信息
    const currentLogPageStart = document.getElementById("currentLogPageStart");
    const currentLogPageEnd = document.getElementById("currentLogPageEnd");
    const totalLogsCount = document.getElementById("totalLogsCount");
    
    if (currentLogPageStart) currentLogPageStart.textContent = this.totalLogs > 0 ? startItem : 0;
    if (currentLogPageEnd) currentLogPageEnd.textContent = this.totalLogs > 0 ? endItem : 0;
    if (totalLogsCount) totalLogsCount.textContent = this.totalLogs;
    
    // 更新按钮状态
    const prevLogPageBtn = document.getElementById("prevLogPageBtn");
    const nextLogPageBtn = document.getElementById("nextLogPageBtn");
    
    if (prevLogPageBtn) {
      prevLogPageBtn.disabled = this.currentLogPage <= 1;
    }
    
    if (nextLogPageBtn) {
      nextLogPageBtn.disabled = this.currentLogPage >= totalPages;
    }
    
    // 生成页码按钮
    this.generateLogPageNumbers(totalPages);
  }

  generateLogPageNumbers(totalPages) {
    const pageNumbersContainer = document.getElementById("logPageNumbers");
    if (!pageNumbersContainer) return;
    
    pageNumbersContainer.innerHTML = '';
    
    if (totalPages <= 1) return;
    
    const maxVisiblePages = 5;
    let startPage = Math.max(1, this.currentLogPage - Math.floor(maxVisiblePages / 2));
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
    
    if (endPage - startPage + 1 < maxVisiblePages) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }
    
    for (let i = startPage; i <= endPage; i++) {
      const pageBtn = document.createElement('button');
      pageBtn.className = `px-3 py-1 border rounded ${
        i === this.currentLogPage 
          ? 'bg-blue-500 text-white border-blue-500' 
          : 'border-gray-300 hover:bg-gray-50'
      }`;
      pageBtn.textContent = i;
      pageBtn.addEventListener('click', () => {
        this.currentLogPage = i;
        this.loadLogsPage();
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
      minute: '2-digit',
      second: '2-digit'
    });
  }

  viewLogDetails(logId) {
    // 这里可以实现查看日志详情的功能
    console.log('查看日志详情:', logId);
    this.app.showAlert('日志详情功能待实现', 'info');
  }

  viewLogError(logId) {
    // 这里可以实现查看错误信息的功能
    console.log('查看错误信息:', logId);
    this.app.showAlert('错误信息查看功能待实现', 'info');
  }
}