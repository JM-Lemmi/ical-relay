function initSelect2(selectorId, enableRedirection = true){
    if(enableRedirection){
    	$(selectorId).on('select2:select', function (event){
    		console.log(event.params.data.id);
    		window.location.href = event.params.data.id;
    	});
    }
	$(selectorId).select2({
		placeholder: "Profil auswählen",
		theme: 'bootstrap4',
		width: 'style',
		//language: 'de'
	});
	$(document).on('select2:open', () => {
		document.querySelector('.select2-search__field').focus();
	});
}