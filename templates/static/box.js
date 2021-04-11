// Classes to manage drawing rectangles on an HTML canvas,
// And drag/resize/remove them
/**
 * TODO List:
 * - Better box labels
 */

/**
 * BoxContainer manages drawing, dragging, resizing rectangles over
 * an image, in an html canvas.
 * 
 * When a box is added an event "box-added" is emitted, and when it gets removed
 * a "box-removed" event is emitted.
 */
class BoxContainer {
    /**
     * 
     * @param {string} id the id of the canvas element
     * @param {string} imgurl The url of the image to use as a canvas background.
     * @param {*} lineOffset The size of the rectangle walls. 4 by default.
     * @returns {BoxContainer}
     */
    constructor(id, imgurl, lineOffset = 4) {
        this.boxes = [];
        this.isCreatingBox = false;
        this.element = document.getElementById(id);
        if (this.element == null) {
            // TODO: error handling
            return;
        }
        this.context = this.element.getContext('2d');
        this.lineOffset = lineOffset;
        this.lineWidth = Math.ceil(lineOffset / 2);
        this.color = 'lightgrey';
        this.selected = new SelectedArea();
        this.mousedown = null;
        // Load and draw image
        this.image = new Image();
        this.image.src = imgurl;
        this.image.onload = this.redraw.bind(this);
        this.registerEvents();
    }

    /**
     * Re-draws the image and the rectangles in the canvas.
     * @returns null
     */
    redraw() {
        if (!this.element) {
            console.log("Cannot redraw boxes");
            return;
        }
        this.context.clearRect(0, 0, this.element.width, this.element.width);
        this.context.drawImage(this.image, 0, 0);
        this.context.beginPath();
        for (var i = 0; i < this.boxes.length; i++) {
            var overrideColor = '';
            var box = this.boxes[i];
            if (i == this.selected.boxId) {
                overrideColor = 'cadetblue';
            }
            box.drawOn(this.context, i + 1, overrideColor);
        }
    }

    /**
     * Get the current box into this.selected, or create a new tmpbox. Then return it.
     * @param {int} x The x coordinate of the point at which to select (optional)
     * @param {int} y The y coordinate of the point at which to select (optional)
     * @returns Box
     */
    getCurrentBox(x = -1, y = -1) {
        // First let's check if we have a selected box. If we do, return that.
        if (this.selected.inBox()) {
            return this.boxes[this.selected.boxId];
        }
        // Now if x or y are not > -1, we want to bail out.
        if (x < 0 || y < 0) {
            return null;
        }
        // Find the position of the mouse relative to the canvas.
        var point = new Point(x, y);
        for (var i = 0; i < this.boxes.length; i++) {
            var position = this.boxes[i].getPosition(point);
            if (position != 'o') {
                this.selected = new SelectedArea(i, position);
                return this.boxes[i]
            }
        }
        // No boxes were found. We're creating a new one. Let's generate it, and make
        // it selected.
        // We add a box that has center at the point and no dimensions.
        this.addBox(point, 0, 0);
        this.isCreatingBox = true;
        // We've selected the new box, and we're dragging it from the bottom right.
        this.selected = new SelectedArea(this.boxes.length - 1, 'se');
        // Now return it.
        return this.boxes[this.selected.boxId];
    };

    /**
     * Add a box of given dimensions at a specific point.
     * @param {Point} center The center of the box to add
     * @param {int} w the width of the box
     * @param {int} l the length of the box
     */
    addBox(center, w, l) {
        var hw = Math.ceil(w / 2);
        var hl = Math.ceil(l / 2);
        var box = new Box(
            center.x - hw, center.y - hl,
            center.x + hw, center.y + hl,
            this.lineWidth, this.lineOffset, this.color);
        this.boxes.push(box);
        let ev = new CustomEvent('box-added', { detail: (this.boxes.length - 1) });
        this.element.dispatchEvent(ev);
        this.redraw();
    }


    /**
     * Register events to handle box resizing.
     */
    registerEvents() {
        // Click the mouse. Create a new box if none is selected.
        // If one is selected, highlight it.
        this.element.onmousedown = this.mouseDown.bind(this);

        // Release the mouse.
        // If we started inside the canvas, we want to 
        // add the temporary box to the image.
        // If not, just clear selection and remove the mousedown
        this.element.onmouseup = this.mouseUp.bind(this);

        // The mouse is moved outside of the canvas.
        // if a temporary box was created, it's discarded.
        this.element.onmouseout = this.mouseOut.bind(this);

        // The mouse is moved inside the element.
        // If a new box is present, the x2,y2 coordinates are changed.
        // If a box is present and it's being dragged, it gets moved.
        // else, it's resized based on which anchor is selected.
        this.element.onmousemove = this.mouseMove.bind(this);

        // Allow removing a box using escape
        window.addEventListener("keydown", this.keyPress.bind(this), false);
    }

    /**
     * Event handler for mouse being clicked.
     * @param {event} e
     * @private
     */
    mouseDown(e) {
        e.preventDefault();
        e.stopPropagation();
        // We don't need relative coordinates here, as we're not using
        // it to draw shapes but just to calculate motion.
        this.mousedown = new Point(e.offsetX, e.offsetY);
        // select the box at the current position.
        this.getCurrentBox(e.offsetX, e.offsetY);
        this.setCursorStyle(e);
        // Redraw to show the selection
        this.redraw();
    }
    /**
     * Event handler for mouse being released.
     * @param {event} e
     * @private
     */
    mouseUp(e) {
        if (this.mousedown == null) { return; }
        e.preventDefault();
        e.stopPropagation();
        this.clearSelection();
        this.setCursorStyle(e);
    }

    /**
     * Event handler for the mouse exiting the canvas.
     * @param {event} e
     * @private
     */
    mouseOut(e) {
        if (this.mousedown == null) { return; }
        e.preventDefault();
        e.stopPropagation();
        // Remove a box that's being created.
        if (this.isCreatingBox) {
            this.removeSelectedBox();
        } else {
            // else just clear the selection.
            this.clearSelection();
        }
    }

    /**
     * Event handler for the mouse moving on the canvas.
     * @param {event} e
     * @private
     */
    mouseMove(e) {
        // We always want to check if we need to change the pointer
        // unless we're in an action.
        if (this.mousedown == null) {
            this.setCursorStyle(e);
            return;
        }
        e.preventDefault();
        e.stopPropagation();
        var position = new Point(e.offsetX, e.offsetY);
        var box = this.getCurrentBox();
        if (box != null) {
            if (this.isCreatingBox) {
                box.x2 = position.x;
                box.y2 = position.y;
            } else {
                var offset = position.sub(this.mousedown);
                box.move(this.selected.position, offset);
            }
        }
        // Reset the previous mouse position to the new value.
        this.mousedown = position;
        this.redraw();
    }

    /**
     * Event handler for keypress.
     * @param {event} e
     * @private
     */
    keyPress(e) {
        if (this.mousedown == null) { return; }
        if (e.key == "Escape") {
            this.removeSelectedBox();
            this.setCursorStyle(e);
        };
    }


    /**
     * Set the style of the mouse cursor.
     * @param {event} e
     * @private
     */
    setCursorStyle(e) {
        // If we've selected a box, we want to return the handle
        // for that box.
        if (this.selected.inBox()) {
            this._setMouseStyle(this.selected.position);
            return;
        }
        var point = new Point(e.offsetX, e.offsetY);
        for (var i = 0; i < this.boxes.length; i++) {
            let position = this.boxes[i].getPosition(point);
            if (position != 'o') {
                this._setMouseStyle(position);
                // We break out as the first box matching will also be
                // the one selected if the user clicks.
                return;
            }
        }
        // We found no match.
        this._setMouseStyle('o');
    }

    /**
     * @param {string} position position identifier
     * @private
     */
    _setMouseStyle(position) {
        var mouseStyle = 'default';
        switch (position) {
            case 'i':
                if (this.mousedown !== null) {
                    mouseStyle = 'move';
                }
                break;
            case 'o':
                break;
            default:
                mouseStyle = position + '-resize';
                break;
        }
        this.element.style.cursor = mouseStyle;
    }

    /**
     * Removes the currently selected box.
     * @private
     */
    removeSelectedBox() {
        if (this.selected.inBox()) {
            this.boxes.splice(this.selected.boxId, 1);
            let ev = new CustomEvent('box-removed', { detail: this.selected.boxId });
            this.element.dispatchEvent(ev);
        }
        this.clearSelection();
    }

    /**
     * Clears all state once we've removed selections.
     * @private
     */
    clearSelection() {
        this.mousedown = null;
        this.selected = new SelectedArea();
        this.isCreatingBox = false;
        this.redraw();
    }
}


/** Class representing a box */
class Box {
    /**
     * Create a box
     * @param {number} x1 the x coordinate of the top vertex
     * @param {number} y1 the y coordinate of the top vertex
     * @param {number} x2 the x coordinate of the bottom vertex
     * @param {number} y2 the y coordinate of the bottom vertex
     * @param {number} lineWidth width of the drawing
     * @param {number} lineOffset the size of the anchors
     * @param {string} color The html color.
     */
    constructor(x1, y1, x2, y2, lineWidth, lineOffset, color) {
        // (x1, y1) must be the top left point,
        // (x2, y2) must be the bottom right point
        // Remember that y increases from the top of the element :)
        this.x1 = x1;
        this.x2 = x2;
        this.y1 = y1;
        this.y2 = y2;
        this.fixCoordinates();
        this.lineOffset = lineOffset;
        this.lineWidth = lineWidth;
        this.color = color;
    }

    /**
     * Return the data for the box, but in terms of center/width/height.
     * @returns {object} A simple k-v representation of the box.
     */
    dimensions() {
        this.fixCoordinates();
        var width = this.x2 - this.x1;
        var length = this.y2 - this.y1;
        return { "x": Math.round(this.center.x), "y": Math.round(this.center.y), "w": width, "l": length };
    }
    /**
     * Draw the box onto a canvas
     * @param {CanvasRenderingContext2D} context the drawing context
     */
    drawOn(context, tag, overrideColor = '') {
        var anchorSize = Math.ceil(this.lineOffset / 2);
        var lo = this.lineOffset;
        function fillRect(x, y) {
            context.fillRect(x - anchorSize, y - anchorSize, lo, lo);
        }
        context.strokeStyle = overrideColor ? overrideColor : this.color;
        context.fillStyle = this.color;
        context.lineWidth = this.lineWidth;
        // Ensure the box is drawable
        this.fixCoordinates();
        context.strokeRect(this.x1, this.y1, (this.x2 - this.x1), (this.y2 - this.y1));
        // Now add the anchors
        fillRect(this.x1, this.y1);
        fillRect(this.x1, this.center.y);
        fillRect(this.x1, this.y2);
        fillRect(this.center.x, this.y1);
        fillRect(this.center.x, this.y2);
        fillRect(this.x2, this.y1);
        fillRect(this.x2, this.center.y);
        fillRect(this.x2, this.y2);
        context.font = "bold 15px sans-serif";
        context.fillText(tag, this.center.x, this.center.y)
    }

    /**
     * Find where is the point with respect to the box.
     * @param {Point} p the point we're evaluating our position against.
     * @returns {string} the position
     */
    getPosition(p) {
        var position = '';
        // Note: this could be made slightly more efficient,
        // at the cost of readability.
        if (this.isLeft(p)) {
            position = 'w';
        } else if (this.isRight(p)) {
            position = 'e';
        }
        var isInHandle = position || this.isXcentered(p);
        // If a match was found, we also want to find the vertical
        // position.
        if (isInHandle) {
            if (this.isTop(p)) { // Top 3 handles
                return 'n' + position;
            } else if (this.isBottom(p)) { // Bottom 3 handles
                return 's' + position;
            }
        }
        if (position && this.isYcentered(p)) { // Left and Right handle
            return position;
        }
        // We're not in a handle, let's check if we're inside or outside the rectangle.
        if (this.isInside(p)) {
            // if we're inside the figure, but nowhere near a handle.
            return 'i';
        }
        return 'o';
    }

    /**
     * 
     * @param {string} position The position we're dragging at
     * @param {Point} offset The offset we're moving for
     */
    move(position, offset) {
        // Dragging
        if (position == 'i') {
            this.x1 += offset.x;
            this.x2 += offset.x;
            this.y1 += offset.y;
            this.y2 += offset.y;
        } else {
            // left or right shift
            if (position.includes('w')) {
                this.x1 += offset.x;
            } else if (position.includes('e')) {
                this.x2 += offset.x;
            }
            // top or bottom shift
            if (position.includes('n')) {
                this.y1 += offset.y;
            } else if (position.includes('s')) {
                this.y2 += offset.y;
            }
        }
        // we mainly want to calculate the new center.
        this.fixCoordinates();
    }
    /** "Private" methods */


    /**
     * Ensures the top-left vertex is in (x1, y1)
     * and the bottom-right one is in (x2, y2)
     * Also calculates the center of the rectangle.
     * @private
     */
    fixCoordinates() {
        if (this.x1 > this.x2) {
            var swap = this.x1;
            this.x1 = this.x2;
            this.x2 = swap;
        }
        if (this.y1 > this.y2) {
            var swap = this.y1;
            this.y1 = this.y2;
            this.y2 = swap;
        }
        this.center = new Point((this.x1 + this.x2) / 2, (this.y1 + this.y2) / 2);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isLeft(p) {
        return (Math.abs(p.x - this.x1) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isRight(p) {
        return (Math.abs(p.x - this.x2) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isTop(p) {
        return (Math.abs(p.y - this.y1) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isBottom(p) {
        return (Math.abs(p.y - this.y2) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isXcentered(p) {
        return (Math.abs(p.x - this.center.x) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isYcentered(p) {
        return (Math.abs(p.y - this.center.y) < this.lineOffset);
    }

    /**
     * @param {Point} p The point to check
     * @returns bool
     * @private
     */
    isInside(p) {
        return (this.x1 - this.lineOffset < p.x &&
            this.x2 + this.lineOffset > p.x &&
            this.y1 - this.lineOffset < p.y &&
            this.y2 + this.lineOffset > p.y);
    }
}

/** Selection management. */
class SelectedArea {
    /**
     * 
     * @param {int} boxId the index of the selected box.
     * @param {string} pos the position at which the box was selected.
     */
    constructor(boxId = -1, pos = 'o') {
        this.boxId = boxId;
        this.position = pos;
    }

    /**
     * Checks if a box is selected.
     * @returns boolean
     */
    inBox() {
        return (this.boxId > -1);
    }
}

/**
 * Point on the canvas.
 */
class Point {
    constructor(x, y) {
        this.x = x;
        this.y = y;
    }

    /**
     * Subtracts points. This works as a vector subtraction.
     * @param {Point} p the point to subtract.
     * @returns 
     */
    sub(p) {
        return new Point(this.x - p.x, this.y - p.y);
    }
}





