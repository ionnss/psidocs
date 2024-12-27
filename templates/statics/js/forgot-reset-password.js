document.querySelector('form').addEventListener('submit', function(e) {
    const chave = document.getElementById('chave').value;
    const confirmarChave = document.getElementById('confirmar_chave').value;

    if (chave !== confirmarChave) {
        e.preventDefault();
        alert('As chaves não coincidem!');
        return;
    }

    const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{10,72}$/;
    if (!regex.test(chave)) {
        e.preventDefault();
        alert('A chave não atende aos requisitos mínimos de segurança!');
        return;
    }
});