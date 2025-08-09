class LLMProxyApp {
    constructor() {
        this.token = localStorage.getItem('token');
        this.apiKey = localStorage.getItem('apiKey');
        this.currentPage = 'dashboard';
        this.allModels = []; // 存储所有模型数据
        this.init();
    }

    init() {
        this.bindEvents();
        this.checkAuth();
    }

    // 通用API调用方法，自动处理401错误
    async apiCall(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json',
                ...options.headers
            }
        };

        const response = await fetch(url, { ...options, headers: defaultOptions.headers });
        
        // 如果遇到401错误，自动跳转到登录页
        if (response.status === 401) {
            this.handleUnauthorized();
            throw new Error('Unauthorized');
        }

        return response;
    }

    // 处理401未授权错误
    handleUnauthorized() {
        this.token = null;
        this.apiKey = null;
        localStorage.removeItem('token');
        localStorage.removeItem('apiKey');
        this.showLoginPage();
        
        // 显示提示信息
        const errorDiv = document.getElementById('loginError');
        errorDiv.textContent = '登录已过期，请重新登录';
        errorDiv.classList.remove('hidden');
    }

    bindEvents() {
        // 登录表单
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.login();
        });

        // 登出按钮
        document.getElementById('logoutBtn').addEventListener('click', () => {
            this.logout();
        });

        // 侧边栏切换
        document.getElementById('sidebarToggle').addEventListener('click', () => {
            this.toggleSidebar();
        });

        // 遮罩层点击
        document.getElementById('overlay').addEventListener('click', () => {
            this.closeSidebar();
        });

        // 导航菜单
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.getAttribute('data-page');
                this.showPage(page);
            });
        });

        // 聊天测试
        document.getElementById('sendMessage').addEventListener('click', () => {
            this.sendChatMessage();
        });

        // 回车发送消息
        document.getElementById('messageInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && e.ctrlKey) {
                this.sendChatMessage();
            }
        });

        // 用户注册表单
        document.getElementById('registerForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.registerUser();
        });

        // 绑定添加用户弹窗
        const addUserBtn = document.getElementById('addUserBtn');
        const addUserModal = document.getElementById('addUserModal');
        const closeModalBtn = document.getElementById('closeModalBtn');
        const cancelBtn = document.getElementById('cancelBtn');

        if (addUserBtn && addUserModal) {
            addUserBtn.addEventListener('click', () => {
                this.showAddUserModal();
            });
        }

        if (closeModalBtn && addUserModal) {
            closeModalBtn.addEventListener('click', () => {
                this.hideAddUserModal();
            });
        }

        if (cancelBtn && addUserModal) {
            cancelBtn.addEventListener('click', () => {
                this.hideAddUserModal();
            });
        }

        // 点击弹窗外部关闭
        if (addUserModal) {
            addUserModal.addEventListener('click', (e) => {
                if (e.target === addUserModal) {
                    this.hideAddUserModal();
                }
            });
        }

        // 模型筛选按钮
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('provider-filter-btn')) {
                this.filterModelsByProvider(e.target.getAttribute('data-provider'));
                
                // 更新按钮样式
                document.querySelectorAll('.provider-filter-btn').forEach(btn => {
                    btn.classList.remove('bg-blue-500', 'text-white');
                    btn.classList.add('bg-gray-200', 'text-gray-700');
                });
                
                e.target.classList.remove('bg-gray-200', 'text-gray-700');
                e.target.classList.add('bg-blue-500', 'text-white');
            }
        });
    }

    async loadUserInfo() {
        try {
            const response = await this.apiCall('/api/user');
            if (response.ok) {
                const data = await response.json();
                this.apiKey = data.user.api_key;
                localStorage.setItem('apiKey', this.apiKey);
            }
        } catch (error) {
            console.error('Failed to load user info:', error);
        }
    }

    async checkAuth() {
        if (this.token) {
            // 如果没有 API Key，尝试加载用户信息
            if (!this.apiKey) {
                await this.loadUserInfo();
            }
            this.showMainApp();
            this.loadDashboard();
        } else {
            this.showLoginPage();
        }
    }

    async login() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const errorDiv = document.getElementById('loginError');

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            const data = await response.json();

            if (response.ok) {
                this.token = data.token;
                localStorage.setItem('token', this.token);
                
                // 获取用户信息，包括 API Key
                await this.loadUserInfo();
                
                this.showMainApp();
                this.loadDashboard();
                errorDiv.classList.add('hidden');
            } else {
                errorDiv.textContent = data.error || '登录失败';
                errorDiv.classList.remove('hidden');
            }
        } catch (error) {
            errorDiv.textContent = '网络错误，请重试';
            errorDiv.classList.remove('hidden');
        }
    }

    async logout() {
        try {
            await fetch('/api/logout', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                },
            });
        } catch (error) {
            console.error('Logout error:', error);
        }

        this.token = null;
        this.apiKey = null;
        localStorage.removeItem('token');
        localStorage.removeItem('apiKey');
        this.showLoginPage();
    }

    showLoginPage() {
        document.getElementById('loginPage').classList.remove('hidden');
        document.getElementById('mainApp').classList.add('hidden');
    }

    showMainApp() {
        document.getElementById('loginPage').classList.add('hidden');
        document.getElementById('mainApp').classList.remove('hidden');
        this.loadModels();
    }

    toggleSidebar() {
        const sidebar = document.getElementById('sidebar');
        const overlay = document.getElementById('overlay');
        
        sidebar.classList.toggle('-translate-x-full');
        overlay.classList.toggle('hidden');
    }

    closeSidebar() {
        const sidebar = document.getElementById('sidebar');
        const overlay = document.getElementById('overlay');
        
        sidebar.classList.add('-translate-x-full');
        overlay.classList.add('hidden');
    }

    showPage(page) {
        // 隐藏所有页面
        document.querySelectorAll('.page-content').forEach(p => {
            p.classList.add('hidden');
        });

        // 显示目标页面
        document.getElementById(`${page}Page`).classList.remove('hidden');

        // 更新导航状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('bg-gray-700', 'text-white');
            item.classList.add('text-gray-300');
        });

        document.querySelector(`[data-page="${page}"]`).classList.add('bg-gray-700', 'text-white');
        document.querySelector(`[data-page="${page}"]`).classList.remove('text-gray-300');

        this.currentPage = page;

        // 加载页面数据
        switch (page) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'logs':
                this.loadLogs();
                break;
            case 'models':
                this.loadModelsPage();
                break;
            case 'users':
                this.loadUsersPage();
                break;
            case 'chat':
                // 聊天页面不需要额外加载
                break;
        }

        // 在移动端关闭侧边栏
        if (window.innerWidth < 1024) {
            this.closeSidebar();
        }
    }

    async loadDashboard() {
        try {
            const response = await this.apiCall('/api/stats');

            if (response.ok) {
                const stats = await response.json();
                this.updateDashboardStats(stats);
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                console.error('Failed to load dashboard:', error);
            }
        }
    }

    updateDashboardStats(stats) {
        document.getElementById('totalRequests').textContent = stats.total_requests || 0;
        document.getElementById('totalTokens').textContent = (stats.total_tokens || 0).toLocaleString();
        document.getElementById('successRate').textContent = `${(stats.success_rate || 0).toFixed(1)}%`;
        document.getElementById('avgDuration').textContent = `${(stats.avg_duration || 0).toFixed(0)}ms`;

        // 更新图表
        this.updateCharts(stats);
    }

    updateCharts(stats) {
        // 模型使用统计
        const modelChart = document.getElementById('modelUsageChart');
        modelChart.innerHTML = '';
        
        if (stats.model_usage) {
            const modelEntries = Object.entries(stats.model_usage);
            if (modelEntries.length > 0) {
                modelEntries.forEach(([model, count]) => {
                    const percentage = (count / stats.total_requests * 100).toFixed(1);
                    const bar = document.createElement('div');
                    bar.className = 'flex items-center mb-2';
                    bar.innerHTML = `
                        <div class="w-20 text-sm text-gray-600">${model}</div>
                        <div class="flex-1 mx-2">
                            <div class="bg-gray-200 rounded-full h-2">
                                <div class="bg-blue-500 h-2 rounded-full" style="width: ${percentage}%"></div>
                            </div>
                        </div>
                        <div class="w-16 text-sm text-gray-600 text-right">${count}</div>
                    `;
                    modelChart.appendChild(bar);
                });
            } else {
                modelChart.innerHTML = '<p class="text-gray-500 text-center">暂无数据</p>';
            }
        }

        // 提供商使用统计
        const providerChart = document.getElementById('providerUsageChart');
        providerChart.innerHTML = '';
        
        if (stats.provider_usage) {
            const providerEntries = Object.entries(stats.provider_usage);
            if (providerEntries.length > 0) {
                providerEntries.forEach(([provider, count]) => {
                    const percentage = (count / stats.total_requests * 100).toFixed(1);
                    const bar = document.createElement('div');
                    bar.className = 'flex items-center mb-2';
                    bar.innerHTML = `
                        <div class="w-20 text-sm text-gray-600">${provider}</div>
                        <div class="flex-1 mx-2">
                            <div class="bg-gray-200 rounded-full h-2">
                                <div class="bg-green-500 h-2 rounded-full" style="width: ${percentage}%"></div>
                            </div>
                        </div>
                        <div class="w-16 text-sm text-gray-600 text-right">${count}</div>
                    `;
                    providerChart.appendChild(bar);
                });
            } else {
                providerChart.innerHTML = '<p class="text-gray-500 text-center">暂无数据</p>';
            }
        }
    }

    async loadModels() {
        try {
            const response = await this.apiCall('/api/models');

            if (response.ok) {
                const data = await response.json();
                this.updateModelSelect(data.data || []);
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                console.error('Failed to load models:', error);
            }
        }
    }

    updateModelSelect(models) {
        const select = document.getElementById('modelSelect');
        select.innerHTML = '<option value="">请选择模型</option>';
        
        // 获取上次选择的模型
        const lastSelectedModel = localStorage.getItem('lastSelectedModel');
        
        models.forEach(model => {
            const option = document.createElement('option');
            option.value = model.id;
            option.textContent = `${model.name} (${model.provider})`;
            
            // 如果是上次选择的模型，设为选中状态
            if (model.id === lastSelectedModel) {
                option.selected = true;
            }
            
            select.appendChild(option);
        });
        
        // 添加模型选择变化监听器，保存选择
        select.addEventListener('change', (e) => {
            if (e.target.value) {
                localStorage.setItem('lastSelectedModel', e.target.value);
            }
        });
    }

    async sendChatMessage() {
        const model = document.getElementById('modelSelect').value;
        const message = document.getElementById('messageInput').value.trim();
        const temperature = parseFloat(document.getElementById('temperature').value);
        const maxTokens = parseInt(document.getElementById('maxTokens').value);

        if (!model) {
            alert('请选择模型');
            return;
        }

        if (!message) {
            alert('请输入消息内容');
            return;
        }

        const sendBtn = document.getElementById('sendMessage');
        const originalText = sendBtn.innerHTML;
        sendBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>发送中...';
        sendBtn.disabled = true;

        try {
            const startTime = Date.now();
            // 使用测试聊天接口，通过 apiCall 方法自动处理 Web Token 认证
            const data = await this.apiCall('/api/test/chat', {
                method: 'POST',
                body: JSON.stringify({
                    model: model,
                    messages: [
                        { role: 'user', content: message }
                    ],
                    temperature: temperature,
                    max_tokens: maxTokens,
                }),
            });

            const duration = Date.now() - startTime;

            if (data) {
                this.showChatResponse(data, duration);
            } else {
                this.showChatError('请求失败');
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                this.showChatError('网络错误：' + error.message);
            }
        } finally {
            sendBtn.innerHTML = originalText;
            sendBtn.disabled = false;
        }
    }

    showChatResponse(data, duration) {
        const responseDiv = document.getElementById('chatResponse');
        const contentDiv = document.getElementById('responseContent');
        const statsDiv = document.getElementById('responseStats');

        if (data.choices && data.choices.length > 0) {
            contentDiv.textContent = data.choices[0].message.content;
        } else {
            contentDiv.textContent = '无响应内容';
        }

        statsDiv.innerHTML = `
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                    <span class="font-medium">模型:</span> ${data.model || 'N/A'}
                </div>
                <div>
                    <span class="font-medium">响应时间:</span> ${duration}ms
                </div>
                <div>
                    <span class="font-medium">Token 使用:</span> ${data.usage?.total_tokens || 'N/A'}
                </div>
                <div>
                    <span class="font-medium">完成原因:</span> ${data.choices?.[0]?.finish_reason || 'N/A'}
                </div>
            </div>
        `;

        responseDiv.classList.remove('hidden');
    }

    showChatError(error) {
        const responseDiv = document.getElementById('chatResponse');
        const contentDiv = document.getElementById('responseContent');
        const statsDiv = document.getElementById('responseStats');

        contentDiv.innerHTML = `<div class="text-red-500"><i class="fas fa-exclamation-triangle mr-2"></i>${error}</div>`;
        statsDiv.innerHTML = '';

        responseDiv.classList.remove('hidden');
    }

    async loadLogs() {
        try {
            const response = await this.apiCall('/api/logs?page=1&page_size=20');

            if (response.ok) {
                const data = await response.json();
                this.updateLogsTable(data);
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                console.error('Failed to load logs:', error);
            }
        }
    }

    updateLogsTable(data) {
        const tableDiv = document.getElementById('logsTable');
        
        if (!data.data || data.data.length === 0) {
            tableDiv.innerHTML = '<p class="text-gray-500 text-center">暂无日志数据</p>';
            return;
        }

        const table = document.createElement('div');
        table.className = 'overflow-x-auto';
        table.innerHTML = `
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
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    ${data.data.map(log => `
                        <tr>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                ${new Date(log.created_at).toLocaleString()}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${log.model}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${log.provider}</td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                                    log.status === 'success' 
                                        ? 'bg-green-100 text-green-800' 
                                        : 'bg-red-100 text-red-800'
                                }">
                                    ${log.status}
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${log.tokens_used || 0}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${log.duration}ms</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${log.client_ip}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        tableDiv.innerHTML = '';
        tableDiv.appendChild(table);
    }

    async loadModelsPage() {
        try {
            const response = await this.apiCall('/api/models');

            if (response.ok) {
                const data = await response.json();
                this.allModels = data.data || []; // 存储所有模型数据
                this.updateModelsTable(this.allModels);
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                console.error('Failed to load models page:', error);
            }
        }
    }

    updateModelsTable(models) {
        const tableDiv = document.getElementById('modelsTable');
        
        if (models.length === 0) {
            tableDiv.innerHTML = '<p class="text-gray-500 text-center">暂无可用模型</p>';
            return;
        }

        const table = document.createElement('div');
        table.className = 'overflow-x-auto';
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
                    ${models.map(model => `
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
                    `).join('')}
                </tbody>
            </table>
        `;

        tableDiv.innerHTML = '';
        tableDiv.appendChild(table);
    }

    showAddUserModal() {
        const modal = document.getElementById('addUserModal');
        modal.classList.remove('hidden');
        
        // 清空表单
        document.getElementById('registerForm').reset();
        document.getElementById('newRequestLimit').value = '100';
        
        // 隐藏错误和成功消息
        document.getElementById('registerError').classList.add('hidden');
        document.getElementById('registerSuccess').classList.add('hidden');
    }

    hideAddUserModal() {
        const modal = document.getElementById('addUserModal');
        modal.classList.add('hidden');
    }

    async registerUser() {
        const username = document.getElementById('newUsername').value;
        const email = document.getElementById('newEmail').value;
        const password = document.getElementById('newPassword').value;
        const requestLimit = parseInt(document.getElementById('newRequestLimit').value);
        const isAdmin = document.getElementById('newIsAdmin').checked;
        
        const errorDiv = document.getElementById('registerError');
        const successDiv = document.getElementById('registerSuccess');

        try {
            const response = await this.apiCall('/api/admin/register', {
                method: 'POST',
                body: JSON.stringify({
                    username,
                    email,
                    password,
                    request_limit: requestLimit,
                    is_admin: isAdmin
                }),
            });

            const data = await response.json();

            if (response.ok) {
                successDiv.textContent = `用户创建成功！API Key: ${data.api_key}`;
                successDiv.classList.remove('hidden');
                errorDiv.classList.add('hidden');
                
                // 重新加载用户列表
                this.loadUsersPage();
                
                // 延迟关闭弹窗
                setTimeout(() => {
                    this.hideAddUserModal();
                }, 2000);
            } else {
                errorDiv.textContent = data.error || '创建用户失败';
                errorDiv.classList.remove('hidden');
                successDiv.classList.add('hidden');
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                errorDiv.textContent = '网络错误，请重试';
                errorDiv.classList.remove('hidden');
                successDiv.classList.add('hidden');
            }
        }
    }

    async loadUsersPage() {
        try {
            const response = await this.apiCall('/api/admin/users');

            if (response.ok) {
                const data = await response.json();
                this.updateUsersTable(data.users || []);
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                console.error('Failed to load users:', error);
            }
        }
    }

    updateUsersTable(users) {
        const tableDiv = document.getElementById('usersTable');
        
        if (users.length === 0) {
            tableDiv.innerHTML = '<p class="text-gray-500 text-center">暂无用户数据</p>';
            return;
        }

        const table = document.createElement('div');
        table.className = 'overflow-x-auto';
        table.innerHTML = `
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">用户名</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">邮箱</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">API Key</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">权限</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">请求统计</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    ${users.map(user => `
                        <tr>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${user.username}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${user.email}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-600">
                                <span class="truncate block w-32" title="${user.api_key}">
                                    ${user.api_key.substring(0, 20)}...
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                                    user.is_active 
                                        ? 'bg-green-100 text-green-800' 
                                        : 'bg-red-100 text-red-800'
                                }">
                                    ${user.is_active ? '活跃' : '禁用'}
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                                    user.is_admin 
                                        ? 'bg-purple-100 text-purple-800' 
                                        : 'bg-gray-100 text-gray-800'
                                }">
                                    ${user.is_admin ? '管理员' : '普通用户'}
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                ${user.total_requests || 0} / ${user.request_limit || 0}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                <button onclick="app.regenerateAPIKey(${user.id})" class="text-blue-600 hover:text-blue-900 mr-3">
                                    重新生成Key
                                </button>
                                <button onclick="app.toggleUserStatus(${user.id}, ${!user.is_active})" class="${user.is_active ? 'text-red-600 hover:text-red-900' : 'text-green-600 hover:text-green-900'}">
                                    ${user.is_active ? '禁用' : '启用'}
                                </button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        tableDiv.innerHTML = '';
        tableDiv.appendChild(table);
    }

    async regenerateAPIKey(userId) {
        if (!confirm('确定要重新生成此用户的 API Key 吗？原 Key 将失效。')) {
            return;
        }

        try {
            const response = await this.apiCall(`/api/admin/users/${userId}/regenerate-key`, {
                method: 'POST',
            });

            if (response.ok) {
                const data = await response.json();
                alert(`新的 API Key: ${data.api_key}`);
                this.loadUsersPage();
            } else {
                const data = await response.json();
                alert('操作失败: ' + (data.error || '未知错误'));
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                alert('网络错误，请重试');
            }
        }
    }

    async toggleUserStatus(userId, enable) {
        const action = enable ? '启用' : '禁用';
        if (!confirm(`确定要${action}此用户吗？`)) {
            return;
        }

        try {
            const response = await this.apiCall(`/api/admin/users/${userId}/toggle`, {
                method: 'POST',
                body: JSON.stringify({ is_active: enable }),
            });

            if (response.ok) {
                alert(`用户${action}成功`);
                this.loadUsersPage();
            } else {
                const data = await response.json();
                alert(`${action}失败: ` + (data.error || '未知错误'));
            }
        } catch (error) {
            if (error.message !== 'Unauthorized') {
                alert('网络错误，请重试');
            }
        }
    }

    // 按提供商筛选模型
    filterModelsByProvider(provider) {
        let filteredModels;
        
        if (provider === 'all') {
            filteredModels = this.allModels;
        } else {
            filteredModels = this.allModels.filter(model => model.provider === provider);
        }
        
        this.updateModelsTable(filteredModels);
    }
}

// 全局变量，用于在HTML中调用方法
let app;

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    app = new LLMProxyApp();
});