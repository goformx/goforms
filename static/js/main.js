// Log viewport dimensions and CSS details
function logViewportDetails() {
    const width = window.innerWidth;
    const height = window.innerHeight;
    console.log(`Viewport size: ${width}px × ${height}px`);

    // Get CSS variables
    const styles = getComputedStyle(document.documentElement);
    const breakpointSm = styles.getPropertyValue('--breakpoint-sm').trim();
    const breakpointMd = styles.getPropertyValue('--breakpoint-md').trim();
    
    console.log('CSS Variables:', {
        '--breakpoint-sm': breakpointSm,
        '--breakpoint-md': breakpointMd
    });

    // Check media queries
    const isAboveSm = window.matchMedia(`(min-width: ${breakpointSm})`).matches;
    const isAboveMd = window.matchMedia(`(min-width: ${breakpointMd})`).matches;
    
    console.log('Media Queries:', {
        'above-sm': isAboveSm,
        'above-md': isAboveMd
    });

    // Check features grid
    const featuresContainer = document.querySelector('.features .container');
    if (featuresContainer) {
        const gridStyle = getComputedStyle(featuresContainer);
        console.log('Features Grid:', {
            'grid-template-columns': gridStyle.gridTemplateColumns,
            'display': gridStyle.display
        });
    }
}

// Log on load and resize
window.addEventListener('load', logViewportDetails);
window.addEventListener('resize', logViewportDetails); 