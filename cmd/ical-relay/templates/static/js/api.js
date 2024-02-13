const apiBase = "/api/";

class API{
    static getHeaders(){
        return {
            "Authorization": localStorage.getItem("token")
        }
    }
    static getJsonHeaders(){
        return {...API.getHeaders(), ...{
            "Content-Type": "application/json"
        }};
    }

    static async get(method, options){
        return handleFetchPromise(
            fetch(apiBase+method+"?"+(new URLSearchParams(options).toString()),{
                cache: "no-store",
                headers: API.getHeaders()
            }),
            method, "GET"
        );
    }

    static async put(method, data){
        return handleFetchPromise(fetch(apiBase+method, {
            method: "PUT",
            headers: API.getJsonHeaders(),
            body: JSON.stringify(data),
        }), method, "PUT")
    }

    static async patch(method, data){
        return handleFetchPromise(fetch(apiBase+method, {
            method: "PATCH",
            headers: API.getJsonHeaders(),
            body: JSON.stringify(data),
        }), method, "PATCH")
    }

    static async delete(method, data){
        return handleFetchPromise(fetch(apiBase+method, {
            method: "DELETE",
            headers: API.getJsonHeaders(),
            body: JSON.stringify(data),
        }), method, "DELETE")
    }

    static async postFormData(method, data) {
        return handleFetchPromise(fetch(apiBase+method, {
            method: "POST",
            headers: API.getHeaders(),
            body: data,
        }), method, "POST")
    }
}

class APIError extends Error{
    constructor(message){
        super(message);
        this.name = 'APIError';
    }
}

function handleFetchPromise(promise, apiMethod, httpMethod){
    return promise.then(async(response) => {
        console.log(response);
        if(!response.ok){
            if(!(response.status >= 400 && response.status < 500)){
                alert("Failed to execute "+httpMethod+" "+apiMethod+": "+await response.text());
            }
            //TODO handle 401, 403 etc.
            throw new APIError; //TODO: possibly make this contain useful info about the API failure
        }
        return response.json()
    }).then(data => {
        return data;
    }).catch((error) => {
        if(error instanceof APIError){
            throw error;
        }
        console.error("Fetch error during "+httpMethod+" "+apiMethod, error);
        if(error instanceof TypeError){
            alert("Netzwerkfehler, bitte Internetverbindung überprüfen!");
        }else{
            alert(error.name+": "+error.message+" while executing "+httpMethod+" "+apiMethod);
        }
        throw error;
    });
}