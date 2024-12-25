document.body.addEventListener("pointermove", (e) => {
    const { currentTarget: el, clientX: x, clientY: y } = e;
    const { top: t, left: l, width: w, height: h } = el.getBoundingClientRect();
    el.style.setProperty('--posX', x - l - w / 1.5);
    el.style.setProperty('--posY', y - t - h / 1.5);
})

document.body.addEventListener('htmx:afterSwap', function(event) {
    // Check if the event target is the content div
    if (event.detail.target.id === 'content') {
        // Add a CSS class to highlight the new content
        event.detail.target.classList.add('highlight');
    }
});