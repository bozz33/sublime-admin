// SublimeGo Dashboard — ApexCharts initialization
// Handles both:
//   1. Dynamic widgets: <div class="sublimego-chart" data-*="..."> (from ChartWidget.Render())
//   2. Named static charts: #lineChart, #donutChart, #barChart (dashboard demo)

// ---------------------------------------------------------------------------
// Theme helpers — reads CSS variables so color changes via Panel.WithPrimaryColor() work
// ---------------------------------------------------------------------------
function isDarkMode() {
    return document.documentElement.classList.contains('dark');
}

function getCSSVar(name, fallback) {
    const val = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
    return val || fallback;
}

function getChartColors() {
    const isDark = isDarkMode();
    return {
        primary:    getCSSVar('--color-primary-500', '#22c55e'),
        primary600: getCSSVar('--color-primary-600', '#16a34a'),
        secondary:  '#3b82f6',
        warning:    '#eab308',
        danger:     '#ef4444',
        text:   isDark ? '#9ca3af' : '#6b7280',
        grid:   isDark ? '#374151' : '#e5e7eb',
        bg:     isDark ? '#1f2937' : '#ffffff'
    };
}

// ---------------------------------------------------------------------------
// Dynamic chart initializer — scans .sublimego-chart data-attributes
// ---------------------------------------------------------------------------
function initDynamicCharts() {
    const els = document.querySelectorAll('.sublimego-chart[data-type]');
    els.forEach(function(el) {
        if (el.dataset._initialized) return;
        el.dataset._initialized = 'true';

        let series, labels, colors;
        try { series = JSON.parse(el.dataset.series || '[]'); } catch(e) { series = []; }
        try { labels  = JSON.parse(el.dataset.labels  || '[]'); } catch(e) { labels = []; }
        try { colors  = JSON.parse(el.dataset.colors  || '[]'); } catch(e) { colors = []; }

        const type   = el.dataset.type   || 'area';
        const height = parseInt(el.dataset.height || '300', 10);
        const themeColors = getChartColors();
        if (!colors.length) { colors = [themeColors.primary, themeColors.secondary, themeColors.warning, themeColors.danger]; }

        const options = {
            series:  series,
            chart:   { type: type, height: height, fontFamily: 'Inter, sans-serif', toolbar: { show: false }, background: 'transparent' },
            colors:  colors,
            labels:  labels,
            stroke:  { curve: 'smooth', width: 2 },
            dataLabels: { enabled: false },
            fill:    { type: 'gradient', gradient: { opacityFrom: 0.35, opacityTo: 0.05, stops: [0, 90, 100] } },
            xaxis:   { categories: labels, axisBorder: { show: false }, axisTicks: { show: false }, labels: { style: { colors: themeColors.text, fontSize: '12px' } } },
            yaxis:   { labels: { style: { colors: themeColors.text, fontSize: '12px' } } },
            grid:    { borderColor: themeColors.grid, strokeDashArray: 4, xaxis: { lines: { show: false } } },
            legend:  { position: 'bottom', labels: { colors: themeColors.text } },
            tooltip: { theme: isDarkMode() ? 'dark' : 'light' },
            plotOptions: {
                pie: { donut: { size: '70%', labels: { show: true, name: { color: themeColors.text }, value: { color: themeColors.text }, total: { show: true, label: 'Total', color: themeColors.text } } } },
                bar: { borderRadius: 4, columnWidth: '55%' }
            }
        };

        const chart = new ApexCharts(el, options);
        chart.render();
    });
}

// ---------------------------------------------------------------------------
// Static named charts — dashboard demo pages (#lineChart, #donutChart, #barChart)
// ---------------------------------------------------------------------------
function initLineChart() {
    const el = document.querySelector('#lineChart');
    if (!el) return null;
    const c = getChartColors();
    const chart = new ApexCharts(el, {
        series: [{ name: 'Offres', data: [180, 220, 250, 280, 310, 350] }, { name: 'Collectes', data: [150, 190, 220, 250, 280, 320] }],
        chart:  { type: 'area', height: 320, fontFamily: 'Inter, sans-serif', toolbar: { show: false }, background: 'transparent' },
        colors: [c.primary, c.secondary],
        fill:   { type: 'gradient', gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0.1, stops: [0, 90, 100] } },
        stroke: { curve: 'smooth', width: 2 },
        dataLabels: { enabled: false },
        xaxis:  { categories: ['Août', 'Sept', 'Oct', 'Nov', 'Déc', 'Jan'], axisBorder: { show: false }, axisTicks: { show: false }, labels: { style: { colors: c.text, fontSize: '12px' } } },
        yaxis:  { labels: { style: { colors: c.text, fontSize: '12px' } } },
        grid:   { borderColor: c.grid, strokeDashArray: 4, xaxis: { lines: { show: false } } },
        legend: { position: 'top', horizontalAlign: 'right', labels: { colors: c.text } },
        tooltip:{ theme: isDarkMode() ? 'dark' : 'light' }
    });
    chart.render();
    return chart;
}

function initDonutChart() {
    const el = document.querySelector('#donutChart');
    if (!el) return null;
    const c = getChartColors();
    const chart = new ApexCharts(el, {
        series: [35, 25, 20, 20],
        chart:  { type: 'donut', height: 256, fontFamily: 'Inter, sans-serif', background: 'transparent' },
        colors: [c.primary, c.secondary, c.warning, c.danger],
        labels: ['Fruits & Légumes', 'Boulangerie', 'Produits Laitiers', 'Plats Préparés'],
        plotOptions: { pie: { donut: { size: '70%', labels: { show: true, name: { show: true, color: c.text }, value: { show: true, color: c.text, formatter: (v) => v + '%' }, total: { show: true, label: 'Total', color: c.text, formatter: () => '100%' } } } } },
        dataLabels: { enabled: false },
        legend:  { show: false },
        stroke:  { width: 0 },
        tooltip: { theme: isDarkMode() ? 'dark' : 'light' }
    });
    chart.render();
    return chart;
}

function initBarChart() {
    const el = document.querySelector('#barChart');
    if (!el) return null;
    const c = getChartColors();
    const chart = new ApexCharts(el, {
        series: [{ name: 'Collectes', data: [44, 55, 57, 56, 61, 58, 63] }, { name: 'Livraisons', data: [35, 41, 36, 26, 45, 48, 52] }],
        chart:  { type: 'bar', height: 350, fontFamily: 'Inter, sans-serif', toolbar: { show: false }, background: 'transparent' },
        colors: [c.primary, c.secondary],
        plotOptions: { bar: { horizontal: false, columnWidth: '55%', borderRadius: 4 } },
        dataLabels: { enabled: false },
        stroke:  { show: true, width: 2, colors: ['transparent'] },
        xaxis:   { categories: ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'], axisBorder: { show: false }, axisTicks: { show: false }, labels: { style: { colors: c.text, fontSize: '12px' } } },
        yaxis:   { labels: { style: { colors: c.text, fontSize: '12px' } } },
        grid:    { borderColor: c.grid, strokeDashArray: 4 },
        legend:  { position: 'top', horizontalAlign: 'right', labels: { colors: c.text } },
        fill:    { opacity: 1 },
        tooltip: { theme: isDarkMode() ? 'dark' : 'light' }
    });
    chart.render();
    return chart;
}

// ---------------------------------------------------------------------------
// Bootstrap on DOMContentLoaded — init both dynamic and static charts
// ---------------------------------------------------------------------------
document.addEventListener('DOMContentLoaded', function() {
    if (typeof ApexCharts === 'undefined') return;
    initDynamicCharts();
    initLineChart();
    initDonutChart();
    initBarChart();
});
