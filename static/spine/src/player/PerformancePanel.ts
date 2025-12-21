/**
 * PerformancePanel - A flexible panel for displaying performance metrics
 * 
 * Features:
 * - FPS counter (averaged over last 60 frames)
 * - GPU frame cost using EXT_disjoint_timer_query_webgl2
 * - Extensible for additional metrics
 * - Color-coded display based on thresholds
 */

export interface PerformancePanelConfig {
	padding: number;
	fontSize: number;
	backgroundColor: string;
	position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
}

export interface MetricConfig {
	label: string;
	getValue: () => number;
	format: (value: number) => string;
	getColor: (value: number) => string;
}

export class PerformancePanel {
	private config: PerformancePanelConfig;
	private metrics: MetricConfig[] = [];
	
	// FPS tracking
	private fpsFrameTimes: number[] = [];
	private fpsLastTime: number = 0;
	private fpsCurrentFPS: number = 0;
	
	// GPU timing tracking
	private gl: WebGLRenderingContext | WebGL2RenderingContext | null = null;
	private timerExt: any = null;
	private queries: any[] = [];
	private queryIndex: number = 0;
	private gpuFrameTime: number = 0;
	private maxQueries: number = 2; // Double buffer queries
	private gpuTimingEnabled: boolean = false;
	
	constructor(config?: Partial<PerformancePanelConfig>) {
		this.config = {
			padding: 10,
			fontSize: 16,
			backgroundColor: 'rgba(0, 0, 0, 0.7)',
			position: 'top-left',
			...config
		};
		
		// Add default metrics
		this.addMetric({
			label: 'FPS',
			getValue: () => this.fpsCurrentFPS,
			format: (value) => Math.round(value).toString(),
			getColor: (value) => {
				if (value >= 50) return '#00ff00';      // Green
				if (value >= 30) return '#ffff00';      // Yellow
				return '#ff0000';                        // Red
			}
		});
		
		this.addMetric({
			label: 'GPU',
			getValue: () => this.gpuFrameTime,
			format: (value) => `${value.toFixed(2)}ms`,
			getColor: (value) => {
				if (value <= 8) return '#00ff00';        // Green (≤8ms = >120fps)
				if (value <= 16) return '#ffff00';       // Yellow (≤16ms = >60fps)
				return '#ff0000';                        // Red (>16ms = <60fps)
			}
		});
	}
	
	/**
	 * Initialize GPU timing support
	 */
	initGPUTiming(gl: WebGLRenderingContext | WebGL2RenderingContext) {
		this.gl = gl;
		
		// Try to get the extension
		const ext = gl.getExtension('EXT_disjoint_timer_query_webgl2');
		if (!ext) {
			console.warn('EXT_disjoint_timer_query_webgl2 not supported');
			return;
		}
		
		this.timerExt = ext;
		this.gpuTimingEnabled = true;
		
		// Create query objects (cast to WebGL2 for createQuery method)
		const gl2 = gl as WebGL2RenderingContext;
		for (let i = 0; i < this.maxQueries; i++) {
			const query = gl2.createQuery();
			if (query) {
				this.queries.push(query);
			}
		}
		
		console.log('GPU timing initialized successfully');
	}
	
	/**
	 * Start GPU timing measurement for current frame
	 */
	beginGPUTiming() {
		if (!this.gpuTimingEnabled || !this.gl || this.queries.length === 0) {
			return;
		}
		
		const gl = this.gl as WebGL2RenderingContext;
		const query = this.queries[this.queryIndex];
		
		// Check if previous query is available
		const available = gl.getQueryParameter(query, gl.QUERY_RESULT_AVAILABLE);
		if (available) {
			// Get result from previous query
			const timeElapsed = gl.getQueryParameter(query, gl.QUERY_RESULT);
			// Convert from nanoseconds to milliseconds
			this.gpuFrameTime = timeElapsed / 1000000;
		}
		
		// Start new query
		gl.beginQuery(this.timerExt.TIME_ELAPSED_EXT, query);
	}
	
	/**
	 * End GPU timing measurement for current frame
	 */
	endGPUTiming() {
		if (!this.gpuTimingEnabled || !this.gl) {
			return;
		}
		
		const gl = this.gl as WebGL2RenderingContext;
		gl.endQuery(this.timerExt.TIME_ELAPSED_EXT);
		
		// Move to next query for double buffering
		this.queryIndex = (this.queryIndex + 1) % this.maxQueries;
	}
	
	/**
	 * Check if GPU timing is disjoint (results invalid)
	 */
	isGPUTimingDisjoint(): boolean {
		if (!this.gpuTimingEnabled || !this.gl) {
			return false;
		}
		
		const gl = this.gl as WebGL2RenderingContext;
		return gl.getParameter(this.timerExt.GPU_DISJOINT_EXT);
	}
	
	/**
	 * Update FPS calculation
	 */
	updateFPS(currentTime: number) {
		if (this.fpsLastTime === 0) {
			this.fpsLastTime = currentTime;
			return;
		}
		
		const deltaTime = currentTime - this.fpsLastTime;
		this.fpsLastTime = currentTime;
		
		// Store frame times for averaging (keep last 60 frames)
		this.fpsFrameTimes.push(deltaTime);
		if (this.fpsFrameTimes.length > 60) {
			this.fpsFrameTimes.shift();
		}
		
		// Calculate average FPS from stored frame times
		if (this.fpsFrameTimes.length > 0) {
			const averageFrameTime = this.fpsFrameTimes.reduce((a, b) => a + b, 0) / this.fpsFrameTimes.length;
			this.fpsCurrentFPS = averageFrameTime > 0 ? 1000 / averageFrameTime : 0;
		}
	}
	
	/**
	 * Add a custom metric to the panel
	 */
	addMetric(metric: MetricConfig) {
		this.metrics.push(metric);
	}
	
	/**
	 * Remove a metric by label
	 */
	removeMetric(label: string) {
		this.metrics = this.metrics.filter(m => m.label !== label);
	}
	
	/**
	 * Draw the performance panel on canvas
	 */
	draw(ctx: CanvasRenderingContext2D, canvasWidth: number, canvasHeight: number) {
		if (this.metrics.length === 0) {
			return;
		}
		
		const { padding, fontSize, backgroundColor, position } = this.config;
		
		// Set font style
		ctx.save();
		ctx.font = `bold ${fontSize}px monospace`;
		ctx.textBaseline = 'top';
		
		// Calculate panel dimensions
		const lineHeight = fontSize * 1.3;
		const lines: Array<{text: string, color: string}> = [];
		let maxWidth = 0;
		
		// Format all metrics
		for (const metric of this.metrics) {
			const value = metric.getValue();
			const formattedValue = metric.format(value);
			const text = `${metric.label}: ${formattedValue}`;
			const color = metric.getColor(value);
			
			const textMetrics = ctx.measureText(text);
			maxWidth = Math.max(maxWidth, textMetrics.width);
			lines.push({ text, color });
		}
		
		const panelWidth = maxWidth + padding * 2;
		const panelHeight = lineHeight * lines.length + padding * 2;
		
		// Calculate position
		let x = padding;
		let y = padding;
		
		switch (position) {
			case 'top-right':
				x = canvasWidth - panelWidth - padding;
				y = padding;
				break;
			case 'bottom-left':
				x = padding;
				y = canvasHeight - panelHeight - padding;
				break;
			case 'bottom-right':
				x = canvasWidth - panelWidth - padding;
				y = canvasHeight - panelHeight - padding;
				break;
			case 'top-left':
			default:
				x = padding;
				y = padding;
				break;
		}
		
		// Draw background
		ctx.fillStyle = backgroundColor;
		ctx.fillRect(x - 5, y - 5, panelWidth, panelHeight);
		
		// Draw metrics
		let currentY = y;
		for (const line of lines) {
			ctx.fillStyle = line.color;
			ctx.fillText(line.text, x, currentY);
			currentY += lineHeight;
		}
		
		ctx.restore();
	}
	
	/**
	 * Get current FPS value
	 */
	getFPS(): number {
		return this.fpsCurrentFPS;
	}
	
	/**
	 * Get current GPU frame time
	 */
	getGPUFrameTime(): number {
		return this.gpuFrameTime;
	}
	
	/**
	 * Check if GPU timing is enabled and working
	 */
	isGPUTimingEnabled(): boolean {
		return this.gpuTimingEnabled;
	}
}
