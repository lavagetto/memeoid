function memeSubmit(container, gif) {
    var toSubmit = { "from": gif, "texts": [], "boxes": [] };
    for (var i = 0; i < container.boxes.length; i++) {
        toSubmit.boxes.push(container.boxes[i].dimensions());
        // TODO: this is horribly brittle.
        toSubmit.texts.push(document.getElementById("text-box-" + i).getElementsByTagName("input")[0].value);
    }
    // TODO: fetch() once the server-side stuff is ready.
    console.log(toSubmit);
}

function memeCheckAndSubmit(container, gif) {
    return function (e) {
        var figure = document.getElementById("meme");
        if (container.boxes.length == 0) {
            e.preventDefault();
            let alertP = document.createElement("p");
            alertP.id = "no-boxes";
            alertP.classList.add("help");
            alertP.classList.add("is-danger");
            let textnode = document.createTextNode("You need to add at least one textarea");
            alertP.appendChild(textnode);
            figure.appendChild(alertP);
        } else {
            let alertP = document.getElementById("no-boxes");
            if (alertP) {
                alertP.remove();
            }
            memeSubmit(container, gif);
            addToForm(container);
        }
    }
}

function addToForm(container) {
    var textBoxContainer = document.getElementById("text-boxes");
    for (var i = 0; i < container.boxes.length; i++) {
        let data = container.boxes[i].dimensions();
        let tid = "box-" + i;
        let input = document.createElement("input");
        input.type = 'hidden';
        input.name = "box";
        input.value = '' + data.x + '|' + data.y + '|' + data.w + '|' + data.l;
        textBoxContainer.appendChild(input);
    }
}

function addTextBox(anchor, id, num) {
    let div = document.createElement("div");
    div.id = id;
    div.classList.add("field");
    let lbl = document.createElement("label");
    lbl.classList.add("label");
    let lblTxt = document.createTextNode("Text area " + (num));
    lbl.appendChild(lblTxt);
    div.appendChild(lbl);
    let input = document.createElement("input");
    input.type = "text";
    input.name = "box-text";
    input.placeholder = "Add text here";
    input.classList.add("input");
    div.appendChild(input);
    anchor.appendChild(div);
}

function removeHook(container) {
    return function (e) {
        // Shift the textarea values, then remove the last
        // text area.
        var id = e.detail;
        var prevElement = document.getElementById("text-box-" + id).getElementsByTagName("input")[0]
        // We still have container.boxes.length + 1 text areas.
        for (let i = id + 1; i <= container.boxes.length; i++) {
            var curElement = document.getElementById("text-box-" + i).getElementsByTagName("input")[0]
            prevElement.value = curElement.value;
            prevElement = curElement;
        }
        document.getElementById("text-box-" + container.boxes.length).remove();
    }
}

var saveHook = function (e) {
    var id = e.detail;
    var textBoxContainer = document.getElementById("text-boxes");
    addTextBox(textBoxContainer, 'text-box-' + id, id + 1);
}


