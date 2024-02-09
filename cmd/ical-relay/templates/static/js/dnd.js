const dropzone = document.getElementById("dropzone");

document.addEventListener("dragenter", function(event) {
    event.preventDefault();
    console.log("dragenter");
    dropzone.style.display = "flex";
});

document.addEventListener("dragover", function(event) {
    event.preventDefault();
    console.log("dragover");
});


dropzone.addEventListener("drop", function(event) {
    event.preventDefault();
    console.log("drop");
    dropzone.style.display = "none";
});


dropzone.addEventListener("drop", function(event) {

    event.preventDefault();
    const { files } = event.dataTransfer;
    handleDroppedFiles(files);

});

async function handleDroppedFiles(files) {
    for (const file of files) {
        const formData = new FormData();

        formData.append('files[]', file, file.name)


        await API.postFormData(
            "profiles/" + localStorage.getItem("profile") + "/newentryfile",
              formData
        ).then((response) => {
            console.log(response);
            if (response == "ok") {
                document.getElementById("import-success").style.display = "block";
            }
        }).catch((error) => {
            console.error(error);
            document.getElementById("import-error").style.display = "block";
        });

    }
}

