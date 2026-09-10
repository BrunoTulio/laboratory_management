// Authentication page JavaScript
// Handles login and register form submissions

document.addEventListener('DOMContentLoaded', function() {
    // Login form
    const loginForm = document.querySelector('.login-form form');
    if (loginForm) {
        loginForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(loginForm);
            const data = Object.fromEntries(formData.entries());
            
            try {
                const response = await fetch('/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok && result.success) {
                    window.location.href = '/dashboard';
                } else {
                    showFlashMessage('Credenciais inválidas', 'error');
                }
            } catch (error) {
                console.error('Login error:', error);
                showFlashMessage('Erro ao fazer login', 'error');
            }
        });
    }
    
    // Register form
    const registerForm = document.querySelector('.register-form form');
    if (registerForm) {
        registerForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(registerForm);
            const data = Object.fromEntries(formData.entries());
            
            // Validate passwords match
            if (data.password !== data.confirm_password) {
                showFlashMessage('As senhas não coincidem', 'error');
                return;
            }
            
            try {
                const response = await fetch('/cadastro', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    showFlashMessage('Cadastro realizado com sucesso! Faça login.', 'success');
                    setTimeout(() => {
                        window.location.href = '/login';
                    }, 2000);
                } else {
                    showFlashMessage('Erro ao cadastrar: ' + (result.error || 'Tente novamente'), 'error');
                }
            } catch (error) {
                console.error('Registration error:', error);
                showFlashMessage('Erro ao cadastrar', 'error');
            }
        });
    }
});

function showFlashMessage(message, type = 'info') {
    const flashDiv = document.createElement('div');
    flashDiv.className = `flash-message ${type}`;
    flashDiv.textContent = message;
    
    const authContent = document.querySelector('.auth-content');
    if (authContent) {
        authContent.insertBefore(flashDiv, authContent.firstChild);
        
        setTimeout(() => {
            flashDiv.remove();
        }, 5000);
    }
}