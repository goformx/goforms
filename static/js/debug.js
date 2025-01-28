// Log viewport dimensions and CSS details
function logViewportDetails() {
    const width = window.innerWidth;
    const height = window.innerHeight;
    console.log(`Viewport size: ${width}px × ${height}px`);

    // Get CSS variables
    const styles = getComputedStyle(document.documentElement);
    const breakpointXs = styles.getPropertyValue('--breakpoint-xs').trim();
    const breakpointSm = styles.getPropertyValue('--breakpoint-sm').trim();
    const breakpointMd = styles.getPropertyValue('--breakpoint-md').trim();
    const breakpointLg = styles.getPropertyValue('--breakpoint-lg').trim();
    const breakpointXl = styles.getPropertyValue('--breakpoint-xl').trim();
    const breakpoint2xl = styles.getPropertyValue('--breakpoint-2xl').trim();
    
    console.log('CSS Breakpoints:', {
        '--breakpoint-xs': breakpointXs,
        '--breakpoint-sm': breakpointSm,
        '--breakpoint-md': breakpointMd,
        '--breakpoint-lg': breakpointLg,
        '--breakpoint-xl': breakpointXl,
        '--breakpoint-2xl': breakpoint2xl
    });

    // Check media queries
    const isAboveXs = window.matchMedia(`(min-width: ${breakpointXs})`).matches;
    const isAboveSm = window.matchMedia(`(min-width: ${breakpointSm})`).matches;
    const isAboveMd = window.matchMedia(`(min-width: ${breakpointMd})`).matches;
    const isAboveLg = window.matchMedia(`(min-width: ${breakpointLg})`).matches;
    const isAboveXl = window.matchMedia(`(min-width: ${breakpointXl})`).matches;
    const isAbove2xl = window.matchMedia(`(min-width: ${breakpoint2xl})`).matches;
    
    console.log('Media Query States:', {
        'above-xs': isAboveXs,
        'above-sm': isAboveSm,
        'above-md': isAboveMd,
        'above-lg': isAboveLg,
        'above-xl': isAboveXl,
        'above-2xl': isAbove2xl
    });

    // Check hero actions
    const heroActions = document.querySelector('.hero-actions');
    if (heroActions) {
        const actionStyle = getComputedStyle(heroActions);
        console.log('Hero Actions:', {
            'display': actionStyle.display,
            'flex-direction': actionStyle.flexDirection,
            'gap': actionStyle.gap,
            'width': heroActions.offsetWidth,
            'padding': actionStyle.padding,
            'margin': actionStyle.margin,
            'justify-content': actionStyle.justifyContent,
            'align-items': actionStyle.alignItems
        });

        // Log individual button widths
        const buttons = heroActions.querySelectorAll('.btn');
        console.log('Hero Buttons:', {
            'count': buttons.length,
            'widths': Array.from(buttons).map(btn => btn.offsetWidth),
            'classes': Array.from(buttons).map(btn => btn.className)
        });
    }

    // Check features grid
    const featuresContainer = document.querySelector('.features .container');
    if (featuresContainer) {
        const gridStyle = getComputedStyle(featuresContainer);
        console.log('Features Grid:', {
            'grid-template-columns': gridStyle.gridTemplateColumns,
            'display': gridStyle.display,
            'gap': gridStyle.gap,
            'padding': gridStyle.padding,
            'width': featuresContainer.offsetWidth,
            'computed-width': gridStyle.width
        });

        // Log individual feature card widths
        const featureCards = featuresContainer.querySelectorAll('.feature-card');
        console.log('Feature Cards:', {
            'count': featureCards.length,
            'widths': Array.from(featureCards).map(card => card.offsetWidth)
        });
    }
}

// Log on load and resize
window.addEventListener('load', logViewportDetails);
window.addEventListener('resize', logViewportDetails); 