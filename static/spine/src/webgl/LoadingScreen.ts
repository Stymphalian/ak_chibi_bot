/******************************************************************************
 * Spine Runtimes License Agreement
 * Last updated January 1, 2020. Replaces all prior versions.
 *
 * Copyright (c) 2013-2020, Esoteric Software LLC
 *
 * Integration of the Spine Runtimes into software or otherwise creating
 * derivative works of the Spine Runtimes is permitted under the terms and
 * conditions of Section 2 of the Spine Editor License Agreement:
 * http://esotericsoftware.com/spine-editor-license
 *
 * Otherwise, it is permitted to integrate the Spine Runtimes into software
 * or otherwise create derivative works of the Spine Runtimes (collectively,
 * "Products"), provided that each user of the Products must obtain their own
 * Spine Editor license and redistribution of the Products in any form must
 * include this license and copyright notice.
 *
 * THE SPINE RUNTIMES ARE PROVIDED BY ESOTERIC SOFTWARE LLC "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL ESOTERIC SOFTWARE LLC BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES,
 * BUSINESS INTERRUPTION, OR LOSS OF USE, DATA, OR PROFITS) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 * THE SPINE RUNTIMES, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *****************************************************************************/

import { TimeKeeper, Color } from "../core/Utils";
import { GLTexture } from "./GLTexture";
import { SceneRenderer, ResizeMode } from "./SceneRenderer";

export class LoadingScreen {
	static FADE_SECONDS = 1;

	private static loaded = 0;
	private static spinnerImg: HTMLImageElement = null;
	private static logoImg: HTMLImageElement = null;

	private renderer: SceneRenderer;
	private logo: GLTexture = null;
	private spinner: GLTexture = null;
	private angle = 0;
	private fadeOut = 0;
	private timeKeeper = new TimeKeeper();
	backgroundColor = new Color(0.135, 0.135, 0.135, 1);
	private tempColor = new Color();
	private firstDraw = 0;

	private static SPINNER_DATA = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAKMAAACjCAYAAADmbK6AAAAACXBIWXMAAAsTAAALEwEAmpwYAAALB2lUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxNDIgNzkuMTYwOTI0LCAyMDE3LzA3LzEzLTAxOjA2OjM5ICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0RXZ0PSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VFdmVudCMiIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczpwaG90b3Nob3A9Imh0dHA6Ly9ucy5hZG9iZS5jb20vcGhvdG9zaG9wLzEuMC8iIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIiB4bWxuczpleGlmPSJodHRwOi8vbnMuYWRvYmUuY29tL2V4aWYvMS4wLyIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNS41IChXaW5kb3dzKSIgeG1wOkNyZWF0ZURhdGU9IjIwMTYtMDktMDhUMTQ6MjU6MTIrMDI6MDAiIHhtcDpNZXRhZGF0YURhdGU9IjIwMTgtMTEtMTVUMTY6NDA6NTkrMDE6MDAiIHhtcDpNb2RpZnlEYXRlPSIyMDE4LTExLTE1VDE2OjQwOjU5KzAxOjAwIiBkYzpmb3JtYXQ9ImltYWdlL3BuZyIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpmZDhlNTljMC02NGJjLTIxNGQtODAyZi1jZDlhODJjM2ZjMGMiIHhtcE1NOkRvY3VtZW50SUQ9ImFkb2JlOmRvY2lkOnBob3Rvc2hvcDpmYmNmZWJlYS03MjY2LWE0NGQtOTI4NS0wOTJmNGNhYzk4ZWEiIHhtcE1NOk9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiIHBob3Rvc2hvcDpDb2xvck1vZGU9IjMiIHRpZmY6T3JpZW50YXRpb249IjEiIHRpZmY6WFJlc29sdXRpb249IjcyMDAwMC8xMDAwMCIgdGlmZjpZUmVzb2x1dGlvbj0iNzIwMDAwLzEwMDAwIiB0aWZmOlJlc29sdXRpb25Vbml0PSIyIiBleGlmOkNvbG9yU3BhY2U9IjY1NTM1IiBleGlmOlBpeGVsWERpbWVuc2lvbj0iMjk3IiBleGlmOlBpeGVsWURpbWVuc2lvbj0iMjQyIj4gPHhtcE1NOkhpc3Rvcnk+IDxyZGY6U2VxPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0iY3JlYXRlZCIgc3RFdnQ6aW5zdGFuY2VJRD0ieG1wLmlpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiIHN0RXZ0OndoZW49IjIwMTYtMDktMDhUMTQ6MjU6MTIrMDI6MDAiIHN0RXZ0OnNvZnR3YXJlQWdlbnQ9IkFkb2JlIFBob3Rvc2hvcCBDQyAyMDE1LjUgKFdpbmRvd3MpIi8+IDxyZGY6bGkgc3RFdnQ6YWN0aW9uPSJzYXZlZCIgc3RFdnQ6aW5zdGFuY2VJRD0ieG1wLmlpZDpiNThlMTlkNi0xYTRjLTQyNDEtODU0ZC01MDVlZjYxMjRhODQiIHN0RXZ0OndoZW49IjIwMTgtMTEtMTVUMTY6NDA6MjMrMDE6MDAiIHN0RXZ0OnNvZnR3YXJlQWdlbnQ9IkFkb2JlIFBob3Rvc2hvcCBDQyAoV2luZG93cykiIHN0RXZ0OmNoYW5nZWQ9Ii8iLz4gPHJkZjpsaSBzdEV2dDphY3Rpb249InNhdmVkIiBzdEV2dDppbnN0YW5jZUlEPSJ4bXAuaWlkOjQ3YzYzYzIwLWJkYjgtYzM0YS1hYzMyLWQ5MDdjOWEyOTA0MCIgc3RFdnQ6d2hlbj0iMjAxOC0xMS0xNVQxNjo0MDo1OSswMTowMCIgc3RFdnQ6c29mdHdhcmVBZ2VudD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgc3RFdnQ6Y2hhbmdlZD0iLyIvPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0iY29udmVydGVkIiBzdEV2dDpwYXJhbWV0ZXJzPSJmcm9tIGFwcGxpY2F0aW9uL3ZuZC5hZG9iZS5waG90b3Nob3AgdG8gaW1hZ2UvcG5nIi8+IDxyZGY6bGkgc3RFdnQ6YWN0aW9uPSJkZXJpdmVkIiBzdEV2dDpwYXJhbWV0ZXJzPSJjb252ZXJ0ZWQgZnJvbSBhcHBsaWNhdGlvbi92bmQuYWRvYmUucGhvdG9zaG9wIHRvIGltYWdlL3BuZyIvPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0ic2F2ZWQiIHN0RXZ0Omluc3RhbmNlSUQ9InhtcC5paWQ6ZmQ4ZTU5YzAtNjRiYy0yMTRkLTgwMmYtY2Q5YTgyYzNmYzBjIiBzdEV2dDp3aGVuPSIyMDE4LTExLTE1VDE2OjQwOjU5KzAxOjAwIiBzdEV2dDpzb2Z0d2FyZUFnZW50PSJBZG9iZSBQaG90b3Nob3AgQ0MgKFdpbmRvd3MpIiBzdEV2dDpjaGFuZ2VkPSIvIi8+IDwvcmRmOlNlcT4gPC94bXBNTTpIaXN0b3J5PiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDo0N2M2M2MyMC1iZGI4LWMzNGEtYWMzMi1kOTA3YzlhMjkwNDAiIHN0UmVmOmRvY3VtZW50SUQ9ImFkb2JlOmRvY2lkOnBob3Rvc2hvcDo2OWRmZjljYy01YzFiLWE5NDctOTc3OS03ODgxZjM0ODk3MDMiIHN0UmVmOm9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiLz4gPHBob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4gPHJkZjpCYWc+IDxyZGY6bGk+eG1wLmRpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2U8L3JkZjpsaT4gPC9yZGY6QmFnPiA8L3Bob3Rvc2hvcDpEb2N1bWVudEFuY2VzdG9ycz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz7qS4aQAAAKZElEQVR42u2de4xVxR3HP8dd3rQryPKo4dGNbtVAQRa1YB93E1tTS7VYqCBiSWhsqGltSx+0xD60tKBorYnNkkBtFUt9xJaGNGlty6EqRAK1KlalshK2C8tzpcIigpz+MbPr5e5y987dM2fv4/tJbjC7v3P2+JvPnTMzZ85MEEURQhQClUpB7gRBAECUYiYwH6gDqoEKoA1oBDYCy4OQJgB92R3yq2S5yRilWASs6CZ0DzA5CNmn/ObOOUpB7kQpRgNLcwj9AHCnMiYZfXIT0C/H2DlRSs0gyeiPaQ6xg4FapUwy+mKUY/wwpUwy+uK4Y/xhpUwy+mKfY3yTUiYZfdHiENsahBxRyiSjL5odYncpXZLRJ3sdYhuVLslYKDKqZpSMBXObVs0oGQumA6OaUTL6Iwg5CBzNMXy7MiYZffNCDjH7g5DdSpVk9M36mGKEZOwxq4Fj3cT8UmmSjEm0Gw8At2UJaQhCtilTeeRWM5EdkmVfOwCIUtQBE4AqILC1ZQuwPgjpSKryWwgy1gfZfjsQ886IKFY2xO9N0jOR69srDOAtzCyYFuCUSrcg6AOcBIYCY4C3gVeT+uNJyvg94GPAxzFjcDuBl4C/AP+UBwXBR4AaYDYwDvgr8Drwi1KScRnwXfut6wNcYT+7Ma97LgX+JRd6jfOAucAXgCvTfl4DvAuMtJVJ0cu41IoYWRHTGWM/1TZmq/2fF8nR14r4U2BQF7+LgMW2k7bY54X4Htr5EvD99s5SlriPArcAY+VGsh1YYDpwMzAgSwy2svhWscpYA/wkx9gKm5S5wBA5kgjnAJcDX7NNpVxcWAZMLUYZJwHDHeKrgXnAdWjZlSS4BLgVuMzRlxt9eeNTxsG2veFyy7gQWAR8Sq54byfeYDssAx3LqLabJldBytgMHMjjuPHAQvTOsU++aJtE/fI4dpevTqZPGV+2veN8+DTwIHCBr29hmVJhJXwA+GAex7cBjxZjm7EFWAL8DfeX39s7NPOy9PKEO7XAV+k8xJYLrcDPgL8Xo4xgJqIuA7bkeXw9ZsBVxMMMYEqex64FfuO7e++bTcAPgD8Bpx2PvRSYKIdi61DOs3edXImAV4Cv2zJsKnYZ24B/AJ+xteRrwAmHBF4mj2JhEnCRg4QnrYh3YZ5NH/J9gUmP5zXYtsdsW+Pl8vffkEex8I5D7HHgGeBhe0dLhKRlbMJM298NXI8Z68rGk8AGeRQLu4DHMGOL2dgJPA78AXguyQvsjScdrTYp2zBDPzfbXl7mmNc64B7MFCbRc/bbfPYHrs343WnbZHsG+BXwZ8y65JS6jOnfwPuBg8BnMQtxjsWsh/0IsNJ2fkR8bAHutbfhG2x7vp9tDzZiFs5/Non2YaHJ2N6OWQf8BxiBeRx4EDPZ9nm544WNVsLtwFWYJ2Wh/fmO3ryw3noHpiv6YyZ5NsuXROhrRypeAv7nfHQJvAOTjbclYuJ3pWcL6YL03rSQjEJIRiEZhZCMQjIKIRmFZBRCMgrJKIRkFJJRCMkoJKMQklFIRiEkoxCSUUhGISSjkIxCSEYhGYWQjEIyCiEZhWQUQjIKySiEZBSSUQjJKCSjEAVCJUAQmCWPoxSjgZuAaZgF348D+zD7ADYDe+2nGWgJQg52dVJvSzOLgqHdmU5ln2IYZou9861Do+x/j8Ss2z7AOrQJWBOEZtetKIrMmt5BEBClWAQsxW3b16OY/QHXA6uD0GzpG0VRPmt6i2KSMeyQrxpYgNl4dCJmV7NcOQEsCULu6ZCR+mAmZiOannAMuC0IWS0Zy0PGKMUCzFZug3p4ullsiJ5obzPOj+H6BgGrohR1KqrSx5bzqhhE7PCvXcY4BZqgoioL4iznunQZq2M8cZXKqSyIs5yr02WsiPHEaiyWSbMxxnNVpMvYFuOJj6mcyoI4y7ktXcbGGE/conIqC+Is58Z0GTfGdNIGzJijKH3W2/KOg43pMi4n//2F92P2KJ4ShCwMQvT4pRwajCFRELIQmGLLf3+ep9pj/TvjCcwI4E5gDp1H0VsxO7k3Zvy7PQjZnXl2DXqXhYydiFKMAcYD44CajH+HZIQfBdYCtwch+854HJh2wkqgFhgGHAaagpAjLhcqGctTxqxOpKgCRgNDMXuK7whCTqU7U9khz3ucAv59xomUe9FVhePGEfs5q1eaQiYKBskoJKMQklFIRiEko5CMQkhGIRmFkIxCMgohGYVkFEIyCskohGQUklEIySiEZBSSUQjJKCSjEJJRSEYhJKOQjEJIRiEZhZCMQjIKIRmFZBSijGXMvIZ+KpZEaF8qeygwHOjb2xdUWQBJqQL6ADOBi4GHMGuGH5Iv3hiG2SJtIWaV4mZgB/AadF6jvVxkvAKzv3UdMNX+bDJm9fx10PV+1qLHIl4P3GLzfh3QBLwKbAZ+DJwuFxkDm5CZmN0Vzsv4/TTMyviVwGOYnRZEPAwBZgDfAC5K+/lo+5kKXAjcBzwPnCz1NuP77LfxO12I2M7FNmFXE+++huVOPfDNDBEz25FzgHuBa4Bzk8x/0jJeCiwCFmP2BsnGh4BbgYFyKDZmZRExnTpbGcywHZySuk0PsbeAG4HZDt+2C6yMb8mjWHgXs+NFd5v09Ac+AYzC7An0EPBKqdSM1wDfBqY7Vvubk263lDhPYHamypVa4MvAHUCq2GvGgcB8YAEwKQ/5nwa33blEVrYDLwJXOhxzLvBJzDhkK/BCMdaMA4C5wF2Y4RrXv7UF+KO9tYh42A08msfoRxVwLfBDYGwxyliLGUMclMexL9rOy075EyvvAKuBlcCbeTa3Pl+MMk7GbP/qyiHg18BWueOFNnu3ymeP8X62h11dbDKm7K3a9Zv7e+BJOeOVRmCNvQO5cgmdt4AueBkH5zCE0FWHpQH4r3zxzlPAw3kcdxg4VmwybnaMfx1YAWxTpyURjtj24wpHuZ7C0yNanzL+FnjZIX4lsEGOJEorcDewKcf4vTb+ZLHJuAeYBxzvJm4/8CPg58AJ+ZE4BzBDNk93k//jwOeAN4qxNw1m5sdV9jZwtlvv48ADujX3GpFtUt0OhPZnJzN63wdtOW7xeSFJPJvehBnBv8/2ricAp2wb8UHgETRvsRDYCiy3IrbPCWi0Mt4BPOf7AoIoivycub5TR/rDmBkjs4Df2fbHJjlQcLwfuNyW13rMXILOkyQ2REUtI5jnnG+mNRFOF3Gh1dlavgozhHUMaLEFGJWImBVnbT4VlYwlSBCYL1iUYgGw6ixhDUHIwo4GmfIrGX3JGKWotj3KbM/cpwQh2yRjYfWmS5EFdD/54ytKk2RMgukxxQjJ2GMm5hAzPEoxRqmSjN6IUgwj9xkr45UxyeiTkQ6x45QuyeiT8x1ia5QuyeiTUaoZJWMxyqiaUTIWzG1aNaNkLJgOzJAoRZVSJhl9McIxfrRSJhl94fq241ClTDL6Yq9jvCYNS0ZvuEwGPopZmlhIRi+sIfeXxtYGIaeUMsnohSCkCViSQ+gezAtOwiW/mvzpkKz3ZnrPxCz1V4dZd6YC8+JSI2YNm+VWXE2ulYyiGPk/nslB8d6ayMkAAAAASUVORK5CYII=";

	private static SPINE_LOGO_DATA = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAKUAAABsCAYAAAALzHKmAAAACXBIWXMAAAsTAAALEwEAmpwYAAALB2lUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxNDIgNzkuMTYwOTI0LCAyMDE3LzA3LzEzLTAxOjA2OjM5ICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0RXZ0PSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VFdmVudCMiIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczpwaG90b3Nob3A9Imh0dHA6Ly9ucy5hZG9iZS5jb20vcGhvdG9zaG9wLzEuMC8iIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIiB4bWxuczpleGlmPSJodHRwOi8vbnMuYWRvYmUuY29tL2V4aWYvMS4wLyIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNS41IChXaW5kb3dzKSIgeG1wOkNyZWF0ZURhdGU9IjIwMTYtMDktMDhUMTQ6MjU6MTIrMDI6MDAiIHhtcDpNZXRhZGF0YURhdGU9IjIwMTgtMTEtMTVUMTY6NDA6NTkrMDE6MDAiIHhtcDpNb2RpZnlEYXRlPSIyMDE4LTExLTE1VDE2OjQwOjU5KzAxOjAwIiBkYzpmb3JtYXQ9ImltYWdlL3BuZyIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDowMTdhZGQ3Ni04OTZlLThlNGUtYmM5MS00ZjEyNjI1YjA3MjgiIHhtcE1NOkRvY3VtZW50SUQ9ImFkb2JlOmRvY2lkOnBob3Rvc2hvcDplMTViNGE2ZS1hMDg3LWEzNDktODdhOS1mNDYzYjE2MzQ0Y2MiIHhtcE1NOk9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiIHBob3Rvc2hvcDpDb2xvck1vZGU9IjMiIHRpZmY6T3JpZW50YXRpb249IjEiIHRpZmY6WFJlc29sdXRpb249IjcyMDAwMC8xMDAwMCIgdGlmZjpZUmVzb2x1dGlvbj0iNzIwMDAwLzEwMDAwIiB0aWZmOlJlc29sdXRpb25Vbml0PSIyIiBleGlmOkNvbG9yU3BhY2U9IjY1NTM1IiBleGlmOlBpeGVsWERpbWVuc2lvbj0iMjk3IiBleGlmOlBpeGVsWURpbWVuc2lvbj0iMjQyIj4gPHhtcE1NOkhpc3Rvcnk+IDxyZGY6U2VxPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0iY3JlYXRlZCIgc3RFdnQ6aW5zdGFuY2VJRD0ieG1wLmlpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiIHN0RXZ0OndoZW49IjIwMTYtMDktMDhUMTQ6MjU6MTIrMDI6MDAiIHN0RXZ0OnNvZnR3YXJlQWdlbnQ9IkFkb2JlIFBob3Rvc2hvcCBDQyAyMDE1LjUgKFdpbmRvd3MpIi8+IDxyZGY6bGkgc3RFdnQ6YWN0aW9uPSJzYXZlZCIgc3RFdnQ6aW5zdGFuY2VJRD0ieG1wLmlpZDpiNThlMTlkNi0xYTRjLTQyNDEtODU0ZC01MDVlZjYxMjRhODQiIHN0RXZ0OndoZW49IjIwMTgtMTEtMTVUMTY6NDA6MjMrMDE6MDAiIHN0RXZ0OnNvZnR3YXJlQWdlbnQ9IkFkb2JlIFBob3Rvc2hvcCBDQyAoV2luZG93cykiIHN0RXZ0OmNoYW5nZWQ9Ii8iLz4gPHJkZjpsaSBzdEV2dDphY3Rpb249InNhdmVkIiBzdEV2dDppbnN0YW5jZUlEPSJ4bXAuaWlkOjJlNjJiMWM2LWIxYzQtNDk0MC04MDMxLWU4ZDkyNTBmODJjNSIgc3RFdnQ6d2hlbj0iMjAxOC0xMS0xNVQxNjo0MDo1OSswMTowMCIgc3RFdnQ6c29mdHdhcmVBZ2VudD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgc3RFdnQ6Y2hhbmdlZD0iLyIvPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0iY29udmVydGVkIiBzdEV2dDpwYXJhbWV0ZXJzPSJmcm9tIGFwcGxpY2F0aW9uL3ZuZC5hZG9iZS5waG90b3Nob3AgdG8gaW1hZ2UvcG5nIi8+IDxyZGY6bGkgc3RFdnQ6YWN0aW9uPSJkZXJpdmVkIiBzdEV2dDpwYXJhbWV0ZXJzPSJjb252ZXJ0ZWQgZnJvbSBhcHBsaWNhdGlvbi92bmQuYWRvYmUucGhvdG9zaG9wIHRvIGltYWdlL3BuZyIvPiA8cmRmOmxpIHN0RXZ0OmFjdGlvbj0ic2F2ZWQiIHN0RXZ0Omluc3RhbmNlSUQ9InhtcC5paWQ6MDE3YWRkNzYtODk2ZS04ZTRlLWJjOTEtNGYxMjYyNWIwNzI4IiBzdEV2dDp3aGVuPSIyMDE4LTExLTE1VDE2OjQwOjU5KzAxOjAwIiBzdEV2dDpzb2Z0d2FyZUFnZW50PSJBZG9iZSBQaG90b3Nob3AgQ0MgKFdpbmRvd3MpIiBzdEV2dDpjaGFuZ2VkPSIvIi8+IDwvcmRmOlNlcT4gPC94bXBNTTpIaXN0b3J5PiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDoyZTYyYjFjNi1iMWM0LTQ5NDAtODAzMS1lOGQ5MjUwZjgyYzUiIHN0UmVmOmRvY3VtZW50SUQ9ImFkb2JlOmRvY2lkOnBob3Rvc2hvcDo2OWRmZjljYy01YzFiLWE5NDctOTc3OS03ODgxZjM0ODk3MDMiIHN0UmVmOm9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDowODMzNWIyYy04NzYyLWQzNGMtOTBhOS02ODJjYjJmYTQ2M2UiLz4gPHBob3Rvc2hvcDpEb2N1bWVudEFu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"

	constructor(renderer: SceneRenderer) {
		this.renderer = renderer;

		this.timeKeeper.maxDelta = 9;

		if (LoadingScreen.logoImg === null) {
			// thank you Apple Inc.
			let isSafari = navigator.userAgent.indexOf("Safari") > -1;

			LoadingScreen.logoImg = new Image();
			LoadingScreen.logoImg.src = LoadingScreen.SPINE_LOGO_DATA;
			if (!isSafari) LoadingScreen.logoImg.crossOrigin = "anonymous";
			LoadingScreen.logoImg.onload = (ev) => {
				LoadingScreen.loaded++;
			}

			LoadingScreen.spinnerImg = new Image();
			LoadingScreen.spinnerImg.src = LoadingScreen.SPINNER_DATA;
			if (!isSafari) LoadingScreen.spinnerImg.crossOrigin = "anonymous";
			LoadingScreen.spinnerImg.onload = (ev) => {
				LoadingScreen.loaded++;
			}
		}
	}

	draw(complete = false) {
		if (complete && this.fadeOut > LoadingScreen.FADE_SECONDS) return;

		this.timeKeeper.update();
		let a = Math.abs(Math.sin(this.timeKeeper.totalTime + 0.75));
		this.angle -= this.timeKeeper.delta / 1.4 * 360 * (1 + 1.5 * Math.pow(a, 5));

		let renderer = this.renderer;
		let canvas = renderer.canvas;
		let gl = renderer.context.gl;

		renderer.resize(ResizeMode.Stretch);

		let cam = renderer.camera;
		// let oldX = cam.position.x;
		// let oldY = cam.position.y;
		// let oldZ = cam.position.z;
		// cam.position.set(canvas.width / 2, canvas.height / 2, 0);
		// cam.viewportWidth = canvas.width;
		// cam.viewportHeight = canvas.height;

		if (!complete) {
			gl.clearColor(this.backgroundColor.r, this.backgroundColor.g, this.backgroundColor.b, this.backgroundColor.a);
			gl.clear(gl.COLOR_BUFFER_BIT);
			this.tempColor.a = 1;
		} else {
			this.fadeOut += this.timeKeeper.delta * (this.timeKeeper.totalTime < 1 ? 2 : 1);
			if (this.fadeOut > LoadingScreen.FADE_SECONDS) {
				// cam.position.set(oldX, oldY, oldZ);
				return;
			}
			a = 1 - this.fadeOut / LoadingScreen.FADE_SECONDS;
			this.tempColor.setFromColor(this.backgroundColor);
			this.tempColor.a = 1 - (a - 1) * (a - 1);
			renderer.begin();
			let x = canvas.width / 2;
			renderer.quad(true,
				0 - x, 0,
				canvas.width - x, 0,
				canvas.width - x, canvas.height,
				0 - x, canvas.height,
				0,
				this.tempColor, this.tempColor, this.tempColor, this.tempColor);
			// renderer.quad(true, 
			// 	0, 0, 
			// 	canvas.width, 0, 
			// 	canvas.width, canvas.height, 
			// 	0, canvas.height,
			// 	this.tempColor, this.tempColor, this.tempColor, this.tempColor);
			renderer.end();
		}
		this.tempColor.set(1, 1, 1, this.tempColor.a);

		if (LoadingScreen.loaded != 2) return;
		if (this.logo === null) {
			this.logo = new GLTexture(renderer.context, LoadingScreen.logoImg);
			this.spinner = new GLTexture(renderer.context, LoadingScreen.spinnerImg);
		}
		this.logo.update(false);
		this.spinner.update(false);

		let logoWidth = this.logo.getImage().width;
		let logoHeight = this.logo.getImage().height;
		let spinnerWidth = this.spinner.getImage().width;
		let spinnerHeight = this.spinner.getImage().height;

		renderer.batcher.setBlendMode(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
		renderer.begin();
		renderer.drawTexture(this.logo,
			// (canvas.width - logoWidth) / 2,
			(-logoWidth) / 2,
			(canvas.height - logoHeight) / 2,
			0.0,
			logoWidth, logoHeight,
			this.tempColor
		);
		renderer.drawTextureRotated(
			this.spinner,
			// (canvas.width - spinnerWidth) / 2, 
			(-spinnerWidth) / 2,
			(canvas.height - spinnerHeight) / 2,
			spinnerWidth, spinnerHeight,
			spinnerWidth / 2, spinnerHeight / 2,
			this.angle,
			this.tempColor
		);
		renderer.end();

		// cam.position.set(oldX, oldY, oldZ);
	}
}
