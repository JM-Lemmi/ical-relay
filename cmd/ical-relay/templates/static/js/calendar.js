function locationToNode(location) {
    try {
        const url = new URL(location);
        const a = document.createElement("a");
        a.href = url;
        a.innerText = url.hostname;
        return a;
    } catch (_) {
        return document.createTextNode(location);
    }
}

function getEventCard(event, show_edit = false, edit_enabled = true) {
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
    if (event.show_start && event.show_end) {
        event_text.innerText = dayjs(event.start).format("HH:mm") + " - " + dayjs(event.end).format("HH:mm");
    } else if (event.show_start) {
        event_text.innerText = "Ab " + dayjs(event.start).format("HH:mm");
    } else if (event.show_end) {
        event_text.innerText = "Bis " + dayjs(event.end).format("HH:mm");
    } else {
        event_text.innerText = "Ganztägig";
    }
    if (event.location) {
        event_text.appendChild(document.createElement("br"));
        event_text.appendChild(locationToNode(event.location));
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
        if (!edit_enabled) {
            edit_button.classList.add("disabled");
        }
        edit_button.addEventListener("click", function (e) {
            e.stopPropagation();
            location.href = event.edit_url + '?' + new URLSearchParams({ 'return-to': window.location.pathname });
        });
        event_body.appendChild(edit_button);
    }
    event_card.appendChild(event_body);
    return event_card;
}

function getDayVStack(date, events, show_edit = false, immutable_past = true) {
    // VStack = Column of events for a day
    // stash || date >= dayjs().startOf("day")
    let day_vstack = document.createElement("div");
    document.createElement("div");
    day_vstack.classList.add("vstack", "col-md-4", "col-xl-2", "pt-2", "day-column", "mb-3");
    let day_title = document.createElement("h5");
    day_title.classList.add("fw-semibold", "text-center", "m-0");
    if (currentType === "month") {
        day_title.role = "button";
        day_title.onclick = function () {
            currentType = "week";
            setSelectedDate(date);
        }
    }
    if (date.isSame(dayjs(), "day")) {
        day_title.classList.add("today");
    }
    day_title.innerText = date.format("dd, DD.MM.YYYY");
    day_vstack.appendChild(day_title);

    let currentMonth = location.hash ? dayjs(location.hash.substring(1)).format("MM") : dayjs().startOf("month").format("MM");
    if (events[date.format("YYYY-MM-DD")] != undefined) {
        let day_events = events[date.format("YYYY-MM-DD")];
        day_events.sort(function (a, b) {
            return dayjs(a.start).diff(dayjs(b.start));
        });
        for (let event of day_events) {
            let edit_enabled = !immutable_past || dayjs(event.start).isAfter(dayjs());
            let card = getEventCard(event, show_edit, edit_enabled);
            if (currentType === "month" && date.format("MM") != currentMonth) {
                card.style.backgroundColor = "#e1e6ea";
            }
            day_vstack.appendChild(card);
        }
    } else {
        let card = getEmptyCard();
        if (currentType === "month" && date.format("MM") != currentMonth) {
            card.style.backgroundColor = "#e1e6ea";
        } else {
            card.classList.add("bg-light");
        }
        day_vstack.appendChild(card);
    }
    return day_vstack;
}

function getEmptyCard() {
    let empty_card = document.createElement("div");
    empty_card.classList.add("card", "text-center", "p-2", "rounded-0");
    empty_card.innerHTML = "&varnothing;";
    return empty_card;
}