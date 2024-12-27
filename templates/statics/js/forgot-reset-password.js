document.addEventListener('DOMContentLoaded', function() {
    // Elementos comuns
    const submitBtn = document.getElementById('submitBtn');
    const spinner = document.getElementById('spinner');
    const submitText = document.getElementById('submitText');
    const toast = document.querySelector('.toast');
    const toastBody = document.querySelector('.toast-body');
    const bsToast = new bootstrap.Toast(toast);

    // Função para mostrar mensagem no toast
    function showMessage(message, success = true) {
        toastBody.textContent = message;
        toast.classList.remove('text-bg-success', 'text-bg-danger');
        toast.classList.add(success ? 'text-bg-success' : 'text-bg-danger');
        bsToast.show();
    }

    // Função para alternar estado do botão
    function toggleButton(loading) {
        spinner.classList.toggle('d-none', !loading);
        submitBtn.disabled = loading;
    }

    // Formulário de recuperação de senha
    const forgotForm = document.getElementById('forgotPasswordForm');
    if (forgotForm) {
        forgotForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            toggleButton(true);

            try {
                const response = await fetch('/forgot-password', {
                    method: 'POST',
                    body: new FormData(this)
                });

                const text = await response.text();
                showMessage(text, response.ok);

                if (response.ok) {
                    forgotForm.reset();
                }
            } catch (error) {
                showMessage('Erro ao processar solicitação. Tente novamente.', false);
            } finally {
                toggleButton(false);
            }
        });
    }

    // Formulário de redefinição de senha
    const resetForm = document.getElementById('resetPasswordForm');
    if (resetForm) {
        resetForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            // Validar senhas
            const chave = document.getElementById('chave').value;
            const confirmarChave = document.getElementById('confirmar_chave').value;

            if (chave !== confirmarChave) {
                showMessage('As chaves não coincidem!', false);
                return;
            }

            const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{10,72}$/;
            if (!regex.test(chave)) {
                showMessage('A chave não atende aos requisitos mínimos de segurança!', false);
                return;
            }

            toggleButton(true);

            try {
                const response = await fetch(window.location.href, {
                    method: 'POST',
                    body: new FormData(this)
                });

                if (response.ok) {
                    window.location.href = '/?msg=senha-alterada';
                } else {
                    const text = await response.text();
                    showMessage(text, false);
                }
            } catch (error) {
                showMessage('Erro ao processar solicitação. Tente novamente.', false);
            } finally {
                toggleButton(false);
            }
        });
    }
});