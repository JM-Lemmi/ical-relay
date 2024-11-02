const dropzone = document.getElementById("dropzone");

document.addEventListener("dragenter", function(event) {
    event.preventDefault();
    dropzone.style.display = "flex";
});

document.addEventListener("dragover", function(event) {
    event.preventDefault();
});


dropzone.addEventListener("drop", function(event) {
    event.preventDefault();
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
       
        await fetch("/api/profiles/"+profileName + "/newentryfile", 
            {
                method: "POST",
                body: formData,
                headers: API.getHeaders(),
            }
        ).then((response) => {
            console.log("response", response);
            if (response.ok) {
                document.getElementById("import-success").style.display = "block";
                return response.json();
            } 
            if (response.status >= 400 && response.status < 500) {
                document.getElementById("import-error").style.display = "block";
                if (response.status == 422) {
                    document.getElementById("import-error").innerText = "Datei konnte nicht verarbeitet werden. Bitte Format überprüfen!";
                }
            }            
        }).catch((error) => {
            if(error instanceof TypeError){
                alert("Netzwerkfehler, bitte Internetverbindung überprüfen!");
            }else{
                console.log(error);
                alert("Failed to execute POST /api/profiles/"+profileName + "/newentryfile: ");
            }
        });
    }
}

