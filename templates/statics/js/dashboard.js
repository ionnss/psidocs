document.addEventListener('DOMContentLoaded', function() {
    // Background animation from index.js
    document.body.addEventListener("pointermove", (e) => {
        const { currentTarget: el, clientX: x, clientY: y } = e;
        const { top: t, left: l, width: w, height: h } = el.getBoundingClientRect();
        el.style.setProperty('--posX', x - l - w / 1.5);
        el.style.setProperty('--posY', y - t - h / 1.5);
    });

    // Sidebar toggle for mobile
    const sidebarToggle = document.createElement('button');
    sidebarToggle.id = 'sidebarCollapse';
    sidebarToggle.className = 'btn btn-dark d-md-none position-fixed';
    sidebarToggle.style.cssText = 'top: 20px; left: 20px; z-index: 1000;';
    sidebarToggle.innerHTML = '<i class="bi bi-list"></i>';
    document.body.appendChild(sidebarToggle);

    sidebarToggle.addEventListener('click', function() {
        document.getElementById('sidebar').classList.toggle('active');
    });

    // Active link handling
    document.querySelectorAll('#sidebar .nav-link').forEach(link => {
        link.addEventListener('click', function() {
            document.querySelectorAll('#sidebar .nav-link').forEach(l => {
                l.parentElement.classList.remove('active');
            });
            this.parentElement.classList.add('active');
        });
    });

    // HTMX after swap handling
    document.body.addEventListener('htmx:afterSwap', function(event) {
        if (event.detail.target.id === 'content-area') {
            // Add animation class
            event.detail.target.classList.add('fade-in');
            // Remove after animation
            setTimeout(() => {
                event.detail.target.classList.remove('fade-in');
            }, 500);
        }
    });

    // Tour guide (opcional)
    if (localStorage.getItem('dashboardTourComplete') !== 'true') {
        // Implementar tour guide aqui
        localStorage.setItem('dashboardTourComplete', 'true');
    }
});

// Função para colapsar o sidebar
function toggleSidebar(collapse = true) {
    const sidebar = document.getElementById('sidebar');
    if (collapse) {
        sidebar.classList.add('active');
    } else {
        sidebar.classList.remove('active');
    }
}

// Observa mudanças no content-area para auto-colapsar no editor de documentos
document.addEventListener('htmx:afterSettle', function(event) {
    if (event.detail.target.id === 'content-area') {
        // Se o conteúdo carregado contém o editor de documentos, colapsa o sidebar
        if (event.detail.target.querySelector('#document-form') || 
            event.detail.target.querySelector('#document-preview')) {
            toggleSidebar(true);
        } else {
            toggleSidebar(false);
        }
    }
}); 