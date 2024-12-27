// Formata CPF enquanto digita (000.000.000-00)
function formatCPF(input) {
    let value = input.value.replace(/\D/g, ''); // Remove tudo que não é dígito
    
    if (value.length > 11) {
        value = value.slice(0, 11);
    }
    
    value = value.replace(/(\d{3})(\d)/, '$1.$2');
    value = value.replace(/(\d{3})(\d)/, '$1.$2');
    value = value.replace(/(\d{3})(\d{1,2})$/, '$1-$2');
    
    input.value = value;
}

// Formata CEP enquanto digita (00000-000)
function formatCEP(input) {
    let value = input.value.replace(/\D/g, '');
    
    if (value.length > 8) {
        value = value.slice(0, 8);
    }
    
    value = value.replace(/(\d{5})(\d)/, '$1-$2');
    
    input.value = value;
}

// Formata telefone enquanto digita
function formatTelefone(input) {
    let value = input.value.replace(/\D/g, '');
    
    if (value.length > 9) {
        value = value.slice(0, 9);
    }
    
    if (value.length > 4) {
        value = value.replace(/(\d{5})(\d)/, '$1-$2');
    }
    
    input.value = value;
}

// Formata DDD (apenas números, máximo 3 dígitos)
function formatDDD(input) {
    let value = input.value.replace(/\D/g, '');
    
    if (value.length > 3) {
        value = value.slice(0, 3);
    }
    
    input.value = value;
} 