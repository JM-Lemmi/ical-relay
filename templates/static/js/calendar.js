function getEventCard(event, show_edit = false) {
    let event_card = document.createElement("div");
    event_card.classList.add("card", "rounded-0");
    let event_body = document.createElement("div");
    event_body.classList.add("card-body");
    let event_title = document.createElement("h6");
    event_title.classList.add("card-title");
    event_title.innerText = event.title;
    event_body.appendChild(event_title);
    let event_text = document.createElement("div");
    event_text.classList.add("card-text");
    event_text.innerText = dayjs(event.start).format("HH:mm") + " - " + dayjs(event.end).format("HH:mm");
    if (event.location) {
        event_text.innerText += "\n" + event.location;
    }
    if (event.description) {
        let description_el = document.createElement("p");
        description_el.classList.add("text-muted", "mb-0");
        description_el.style.whiteSpace = "nowrap";
        description_el.style.overflow = "hidden";
        description_el.style.textOverflow = "ellipsis";
        description_el.innerText = event.description;
        let description_visible = false;
        description_el.addEventListener("click", function (e) {
            e.stopPropagation();
            if (description_visible) {
                description_el.style.whiteSpace = "nowrap";
                description_el.style.overflow = "hidden";
                description_el.style.textOverflow = "ellipsis";
                description_visible = false;
            } else {
                description_el.style.whiteSpace = "normal";
                description_el.style.overflow = "visible";
                description_el.style.textOverflow = "visible";
                description_visible = true;
            }
        }); 
        event_text.appendChild(description_el);
    }
    event_body.appendChild(event_text);
    if (show_edit) {
        let edit_button = document.createElement("button");
        edit_button.classList.add("btn", "btn-sm", "btn-outline-secondary", "rounded-circle", "edit-button");
        edit_button.addEventListener("click", function (e) {
            e.stopPropagation();
            location.href = event.edit_url + '?' + new URLSearchParams({'return-to': window.location.pathname});
        });
        event_body.appendChild(edit_button);
    }
    event_card.appendChild(event_body);
    return event_card;
}

function getDayVStack(date, events, show_edit = false) {
    let day_vstack = document.createElement("div");
    document.createElement("div");
    day_vstack.classList.add("vstack", "col-md-4", "col-xl-2", "pt-2", "day-column", "mb-3");
    let day_title = document.createElement("h5");
    day_title.classList.add("fw-semibold", "text-center", "m-0");
    day_title.innerText = date.format("dd, DD.MM.YYYY");
    day_vstack.appendChild(day_title);
    if (events[date.format("YYYY-MM-DD")] != undefined) {
        for (let event of events[date.format("YYYY-MM-DD")]) {
            day_vstack.appendChild(getEventCard(event, show_edit));
        }
    } else {
        day_vstack.appendChild(getEmptyCard());
    }
    if (date >= date.add(1, "month")) {
        day_vstack.classList.add("bg-light");
    }
    return day_vstack;
}

function getEmptyCard() {
    let empty_card = document.createElement("div");
    empty_card.classList.add("card", "bg-light", "text-center", "p-2", "rounded-0");
    empty_card.innerHTML = "&varnothing;";
    return empty_card;
}