var ChangeImage = function() {
  const file = $('#image').prop('files')[0];
  toBase64(file).then(onConverted());
}

var toBase64 = function(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = error => reject(error);
  });
}

var onConverted = function() {
  return function(v) {
    App.imgdata = v;
    $('#preview').attr('src', v);
  }
}

var SubmitForm = function() {
  if (!App.imgdata) {
    $("#warning").text("Image is Empty").removeClass("hidden").addClass("visible");
    return false;
  }
  $("#submit").addClass('disabled');
  var action  = $("#action").val();
  var image = App.imgdata;
  const data = {action, image};
  request(data, (res)=>{
    $("#result").text(res.message);
    $("#info").removeClass("hidden").addClass("visible");
    ScrollBottom();
  }, (e)=>{
    console.log(e.responseJSON.message);
    $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
    $("#submit").removeClass('disabled');
  });
};

var request = function(data, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};

var ScrollBottom = function() {
  var bottom = document.documentElement.scrollHeight - document.documentElement.clientHeight;
  window.scroll(0, bottom);
}

var App = { imgdata: null, url: location.origin + {{ .ApiPath }} };
