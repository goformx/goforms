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

    // Debug Newsletter Grid
    const newsletterGrid = document.querySelector('.newsletter-grid');
    if (newsletterGrid) {
        const gridStyle = getComputedStyle(newsletterGrid);
        console.log('Newsletter Grid:', {
            'display': gridStyle.display,
            'grid-template-columns': gridStyle.gridTemplateColumns,
            'gap': gridStyle.gap,
            'width': newsletterGrid.offsetWidth,
            'computed-width': gridStyle.width,
            'position': gridStyle.position,
            'margin': gridStyle.margin,
            'padding': gridStyle.padding
        });

        // Log individual section widths and styles
        const sections = newsletterGrid.querySelectorAll('section');
        console.log('Newsletter Sections:', {
            'count': sections.length,
            'widths': Array.from(sections).map(section => section.offsetWidth),
            'classes': Array.from(sections).map(section => section.className)
        });

        // Log container details
        const container = document.querySelector('.newsletter-layout .container');
        if (container) {
            const containerStyle = getComputedStyle(container);
            console.log('Newsletter Container:', {
                'width': container.offsetWidth,
                'computed-width': containerStyle.width,
                'max-width': containerStyle.maxWidth,
                'margin': containerStyle.margin,
                'padding': containerStyle.padding
            });
        }

        // Log parent layout details
        const layout = document.querySelector('.newsletter-layout');
        if (layout) {
            const layoutStyle = getComputedStyle(layout);
            console.log('Newsletter Layout:', {
                'width': layout.offsetWidth,
                'computed-width': layoutStyle.width,
                'padding': layoutStyle.padding,
                'margin': layoutStyle.margin,
                'background': layoutStyle.background
            });
        }

        // Check for any overflowing content
        const isHorizontallyOverflowing = newsletterGrid.scrollWidth > newsletterGrid.clientWidth;
        const isVerticallyOverflowing = newsletterGrid.scrollHeight > newsletterGrid.clientHeight;
        console.log('Newsletter Grid Overflow:', {
            'horizontal': isHorizontallyOverflowing,
            'vertical': isVerticallyOverflowing,
            'scrollWidth': newsletterGrid.scrollWidth,
            'clientWidth': newsletterGrid.clientWidth,
            'scrollHeight': newsletterGrid.scrollHeight,
            'clientHeight': newsletterGrid.clientHeight
        });

        // Log any CSS animations/transitions
        console.log('Newsletter Grid Transitions:', {
            'transition': gridStyle.transition,
            'animation': gridStyle.animation,
            'transform': gridStyle.transform
        });
    }

    // Check hero actions
    const heroActions = document.querySelector('.hero-actions');
    if (heroActions) {
        const computedStyle = window.getComputedStyle(heroActions);
        console.log('Hero Actions Computed:', {
            display: computedStyle.display,
            flexDirection: computedStyle.flexDirection,
            gap: computedStyle.gap,
            width: heroActions.offsetWidth,
            padding: computedStyle.padding,
            margin: computedStyle.margin,
            justifyContent: computedStyle.justifyContent,
            alignItems: computedStyle.alignItems
        });

        // Get all stylesheets affecting hero-actions
        const sheets = document.styleSheets;
        console.log('Stylesheets affecting hero-actions:');
        for (let sheet of sheets) {
            try {
                const rules = sheet.cssRules || sheet.rules;
                for (let rule of rules) {
                    if (rule.selectorText && rule.selectorText.includes('hero-actions')) {
                        console.log('Rule from:', sheet.href || 'inline', {
                            selector: rule.selectorText,
                            styles: rule.style.cssText
                        });
                    }
                }
            } catch (e) {
                console.log('Could not read stylesheet:', sheet.href);
            }
        }

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

    // Debug Demo Grid
    const demoGrid = document.querySelector('.demo-grid');
    if (demoGrid) {
        const gridStyle = getComputedStyle(demoGrid);
        console.log('Demo Grid:', {
            'display': gridStyle.display,
            'grid-template-columns': gridStyle.gridTemplateColumns,
            'gap': gridStyle.gap,
            'width': demoGrid.offsetWidth,
            'height': demoGrid.offsetHeight,
            'padding': gridStyle.padding,
            'margin': gridStyle.margin,
            'border': gridStyle.border
        });

        const sections = demoGrid.querySelectorAll('section');
        console.log('Demo Sections:', {
            'count': sections.length,
            'first': sections[0]?.getBoundingClientRect(),
            'last': sections[sections.length - 1]?.getBoundingClientRect()
        });

        const container = document.querySelector('.demo-layout .container');
        if (container) {
            console.log('Demo Container:', {
                'width': container.offsetWidth,
                'height': container.offsetHeight,
                'padding': getComputedStyle(container).padding,
                'margin': getComputedStyle(container).margin,
                'border': getComputedStyle(container).border
            });
        }

        const layout = document.querySelector('.demo-layout');
        if (layout) {
            console.log('Demo Layout:', {
                'width': layout.offsetWidth,
                'height': layout.offsetHeight,
                'padding': getComputedStyle(layout).padding,
                'margin': getComputedStyle(layout).margin,
                'border': getComputedStyle(layout).border
            });
        }

        const isHorizontallyOverflowing = demoGrid.scrollWidth > demoGrid.clientWidth;
        const isVerticallyOverflowing = demoGrid.scrollHeight > demoGrid.clientHeight;
        console.log('Demo Grid Overflow:', {
            'isHorizontallyOverflowing': isHorizontallyOverflowing,
            'isVerticallyOverflowing': isVerticallyOverflowing,
            'scrollWidth': demoGrid.scrollWidth,
            'clientWidth': demoGrid.clientWidth,
            'scrollHeight': demoGrid.scrollHeight,
            'clientHeight': demoGrid.clientHeight
        });

        console.log('Demo Grid Transitions:', {
            'transition': gridStyle.transition,
            'transform': gridStyle.transform
        });
    }
}

// Log on load and resize
window.addEventListener('load', logViewportDetails);
window.addEventListener('resize', logViewportDetails);

// Debug information for layout and styling
document.addEventListener('DOMContentLoaded', () => {
    // Viewport size
    console.log('Viewport size:', `${window.innerWidth}px × ${window.innerHeight}px`);

    // CSS Breakpoints
    const breakpoints = {
        '--breakpoint-xs': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-xs'),
        '--breakpoint-sm': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-sm'),
        '--breakpoint-md': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-md'),
        '--breakpoint-lg': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-lg'),
        '--breakpoint-xl': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-xl'),
        '--breakpoint-2xl': getComputedStyle(document.documentElement).getPropertyValue('--breakpoint-2xl')
    };
    console.log('CSS Breakpoints:', breakpoints);

    // Media Query States
    const mediaStates = {
        'above-xs': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-xs']})`).matches,
        'above-sm': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-sm']})`).matches,
        'above-md': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-md']})`).matches,
        'above-lg': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-lg']})`).matches,
        'above-xl': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-xl']})`).matches,
        'above-2xl': window.matchMedia(`(min-width: ${breakpoints['--breakpoint-2xl']})`).matches
    };
    console.log('Media Query States:', mediaStates);

    // Grid Container
    const gridContainer = document.querySelector('.grid-container');
    if (gridContainer) {
        const gridStyles = getComputedStyle(gridContainer);
        console.log('Grid Container:', {
            display: gridStyles.display,
            'grid-template-columns': gridStyles.gridTemplateColumns,
            gap: gridStyles.gap,
            width: gridContainer.offsetWidth,
            'computed-width': gridStyles.width,
            'max-width': gridStyles.maxWidth,
            margin: gridStyles.margin,
            padding: gridStyles.padding,
            'box-sizing': gridStyles.boxSizing
        });

        // Grid Items
        const gridItems = gridContainer.querySelectorAll('.grid-item');
        console.log('Grid Items:', {
            count: gridItems.length,
            widths: Array.from(gridItems).map(item => item.offsetWidth),
            heights: Array.from(gridItems).map(item => item.offsetHeight),
            classes: Array.from(gridItems).map(item => Array.from(item.classList).filter(c => c !== 'grid-item')),
            'min-widths': Array.from(gridItems).map(item => getComputedStyle(item).minWidth),
            'box-sizing': Array.from(gridItems).map(item => getComputedStyle(item).boxSizing)
        });

        // Grid Layout
        console.log('Grid Layout:', {
            columns: gridStyles.gridTemplateColumns.split(' ').length,
            'column-widths': gridStyles.gridTemplateColumns.split(' '),
            'row-heights': gridStyles.gridTemplateRows.split(' '),
            'auto-rows': gridStyles.gridAutoRows,
            'auto-columns': gridStyles.gridAutoColumns,
            'grid-areas': gridStyles.gridTemplateAreas
        });

        // Grid Overflow
        console.log('Grid Overflow:', {
            horizontal: gridContainer.scrollWidth > gridContainer.clientWidth,
            vertical: gridContainer.scrollHeight > gridContainer.clientHeight,
            scrollWidth: gridContainer.scrollWidth,
            clientWidth: gridContainer.clientWidth,
            scrollHeight: gridContainer.scrollHeight,
            clientHeight: gridContainer.clientHeight,
            'overflow-x': gridStyles.overflowX,
            'overflow-y': gridStyles.overflowY
        });

        // Grid Transitions
        console.log('Grid Transitions:', {
            transition: gridStyles.transition,
            animation: gridStyles.animation,
            transform: gridStyles.transform
        });
    }
}); 