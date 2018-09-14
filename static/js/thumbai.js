function markFieldError(err) {
    $('#'+err.Name).addClass('is-invalid');
    $('#'+err.Name+'Error').html(err.Message);
}

function markFieldErrors(errs) {
    for (var i in errs) {
        markFieldError(errs[i]);
    }
}

function applyOnKeydownRemoveError(formId) {
    $('#'+formId).find('.form-control').each(function(){
        $(this).keydown(function(){
            $(this).removeClass('is-invalid');
        });
    })
}

function disableWithSpinner(id) {
    var obj = $('#'+id);
    var text = obj.html();
    obj.data('previous-text', text);
    obj.prop('disabled', true).html('<i class="fas fa-spinner fa-pulse"></i> ' + text);
}

function enableWithoutSpinner(id) {
    var obj = $('#'+id);
    obj.prop('disabled', false).html(obj.data('previous-text'));
}