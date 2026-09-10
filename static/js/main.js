// Main JavaScript file for laboratory management system
// This provides basic functionality - students will enhance this

// CSRF token handling
function getCSRFToken() {
    const metaTag = document.querySelector('meta[name="csrf-token"]');
    return metaTag ? metaTag.getAttribute('content') : '';
}

// Flash message handling
function showFlashMessage(message, type = 'info') {
    const flashDiv = document.createElement('div');
    flashDiv.className = `flash-message ${type}`;
    flashDiv.textContent = message;
    
    const mainContent = document.querySelector('.main-content');
    if (mainContent) {
        mainContent.insertBefore(flashDiv, mainContent.firstChild);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            flashDiv.remove();
        }, 5000);
    }
}

// API helper functions
async function apiRequest(url, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': getCSRFToken()
        }
    };
    
    const mergedOptions = { ...defaultOptions, ...options };
    
    try {
        const response = await fetch(url, mergedOptions);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API request failed:', error);
        showFlashMessage('Erro ao processar solicitação', 'error');
        throw error;
    }
}

// Booking functions
async function approveBooking(bookingId) {
    if (!confirm('Tem certeza que deseja aprovar este agendamento?')) {
        return;
    }
    
    try {
        await apiRequest('/admin/agendamentos/aprovar', {
            method: 'POST',
            body: JSON.stringify({ booking_id: bookingId })
        });
        
        showFlashMessage('Agendamento aprovado com sucesso', 'success');
        location.reload();
    } catch (error) {
        // Error already handled in apiRequest
    }
}

async function rejectBooking(bookingId) {
    const reason = prompt('Motivo da rejeição:');
    if (!reason) {
        return;
    }
    
    try {
        await apiRequest('/admin/agendamentos/rejeitar', {
            method: 'POST',
            body: JSON.stringify({ 
                booking_id: bookingId,
                rejection_reason: reason 
            })
        });
        
        showFlashMessage('Agendamento rejeitado com sucesso', 'success');
        location.reload();
    } catch (error) {
        // Error already handled in apiRequest
    }
}

async function cancelBooking(bookingId) {
    const reason = prompt('Motivo do cancelamento:');
    if (!reason) {
        return;
    }
    
    try {
        await apiRequest('/agendamentos/cancelar', {
            method: 'POST',
            body: JSON.stringify({ 
                booking_id: bookingId,
                cancellation_reason: reason 
            })
        });
        
        showFlashMessage('Agendamento cancelado com sucesso', 'success');
        location.reload();
    } catch (error) {
        // Error already handled in apiRequest
    }
}

// Notification functions
async function markAsRead(notificationId) {
    try {
        await apiRequest('/notificacoes/marcar-lida', {
            method: 'POST',
            body: JSON.stringify({ notification_id: notificationId })
        });
        
        // Update UI
        const notificationItem = document.querySelector(`[data-notification-id="${notificationId}"]`);
        if (notificationItem) {
            notificationItem.classList.remove('unread');
        }
        
        // Update unread count
        updateUnreadCount();
    } catch (error) {
        // Error already handled in apiRequest
    }
}

async function markAllAsRead() {
    try {
        await apiRequest('/notificacoes/marcar-todas-lidas', {
            method: 'POST'
        });
        
        showFlashMessage('Todas as notificações marcadas como lidas', 'success');
        location.reload();
    } catch (error) {
        // Error already handled in apiRequest
    }
}

async function updateUnreadCount() {
    try {
        const data = await apiRequest('/notificacoes/nao-lidas');
        const badge = document.querySelector('.notification-badge');
        if (badge) {
            if (data.unread_count > 0) {
                badge.textContent = data.unread_count;
                badge.style.display = 'inline';
            } else {
                badge.style.display = 'none';
            }
        }
    } catch (error) {
        console.error('Failed to update unread count:', error);
    }
}

// Availability checking
async function checkAvailability() {
    const laboratoryId = document.getElementById('laboratory_id').value;
    const date = document.getElementById('booking_date').value;
    const startTime = document.getElementById('start_time').value;
    const endTime = document.getElementById('end_time').value;
    
    if (!laboratoryId || !date || !startTime || !endTime) {
        showFlashMessage('Preencha todos os campos para verificar disponibilidade', 'warning');
        return;
    }
    
    try {
        const data = await apiRequest(`/disponibilidade/verificar?laboratory_id=${laboratoryId}&date=${date}&start_time=${startTime}&end_time=${endTime}`);
        
        const resultDiv = document.getElementById('availability_result');
        if (data.available) {
            resultDiv.innerHTML = '<p class="available">✓ Laboratório disponível para este horário</p>';
            resultDiv.className = 'availability-result available';
        } else {
            resultDiv.innerHTML = '<p class="unavailable">✗ Laboratório não disponível para este horário</p>';
            resultDiv.className = 'availability-result unavailable';
        }
    } catch (error) {
        // Error already handled in apiRequest
    }
}

// Filter functions
function applyFilters() {
    const form = document.querySelector('.filters');
    if (form) {
        const formData = new FormData(form);
        const params = new URLSearchParams(formData);
        window.location.href = `${window.location.pathname}?${params.toString()}`;
    }
}

function clearFilters() {
    window.location.href = window.location.pathname;
}

// Tab switching
function switchTab(tabId) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(`${tabId}-tab`).classList.add('active');
    
    // Add active class to selected tab button
    document.querySelector(`[data-tab="${tabId}"]`).classList.add('active');
}

// View switching for agenda
function switchView(view) {
    // Hide all views
    document.querySelectorAll('.day-view, .week-view, .month-view').forEach(view => {
        view.style.display = 'none';
    });
    
    // Remove active class from all view buttons
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected view
    document.querySelector(`.${view}-view`).style.display = 'block';
    
    // Add active class to selected view button
    document.querySelector(`[data-view="${view}"]`).classList.add('active');
}

// Event listeners setup
document.addEventListener('DOMContentLoaded', function() {
    // Mark all as read button
    const markAllReadBtn = document.getElementById('mark_all_read');
    if (markAllReadBtn) {
        markAllReadBtn.addEventListener('click', markAllAsRead);
    }
    
    // Check availability button
    const checkAvailabilityBtn = document.getElementById('check_availability');
    if (checkAvailabilityBtn) {
        checkAvailabilityBtn.addEventListener('click', checkAvailability);
    }
    
    // Filter buttons
    const applyFiltersBtn = document.getElementById('apply_filters');
    if (applyFiltersBtn) {
        applyFiltersBtn.addEventListener('click', applyFilters);
    }
    
    const clearFiltersBtn = document.getElementById('clear_filters');
    if (clearFiltersBtn) {
        clearFiltersBtn.addEventListener('click', clearFilters);
    }
    
    // Tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            switchTab(this.dataset.tab);
        });
    });
    
    // View buttons
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            switchView(this.dataset.view);
        });
    });
    
    // Initialize unread count
    updateUnreadCount();
});