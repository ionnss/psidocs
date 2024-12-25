// Easter Egg
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded, initializing easter egg...');
    
    const logo = document.getElementById('footerLogo');
    const easterEggText = document.getElementById('easterEggText');
    
    if (!logo || !easterEggText) {
        console.error('Elements not found:', { logo, easterEggText });
        return;
    }

    logo.addEventListener('click', function() {
        console.log('Logo clicked');
        
        // Rotate logo
        this.style.transform = 'rotate(360deg)';
        setTimeout(() => {
            this.style.transform = 'none';
        }, 500);

        // Toggle message
        easterEggText.classList.toggle('show');
    });
});

// Background animation
document.body.addEventListener("pointermove", (e) => {
    const { currentTarget: el, clientX: x, clientY: y } = e;
    const { top: t, left: l, width: w, height: h } = el.getBoundingClientRect();
    el.style.setProperty('--posX', x - l - w / 1.5);
    el.style.setProperty('--posY', y - t - h / 1.5);
});

document.body.addEventListener('htmx:afterSwap', function(event) {
    // Check if the event target is the content div
    if (event.detail.target.id === 'content') {
        // Add a CSS class to highlight the new content
        event.detail.target.classList.add('highlight');
    }
});