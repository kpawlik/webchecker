/*
Web checker javascript library
*/
$(window).load(function(){
	
	$("#busyDialog").dialog({  autoOpen: false,    
        dialogClass: "loadingScreenWindow",
		title: "",
        closeOnEscape: false,
        draggable: false,
		resizable: false,
		modal: true,
		width: 300, height: 50});
	$(document).ajaxStart(function() {
	   $("#busyDialog" ).dialog("open");
 	}).ajaxStop(function() {
		$("#busyDialog" ).dialog("close");
 	});
	$( "#addDialog" ).dialog({autoOpen: false,
	 							modal: true,
								title: "Add new record",
								width: 600,
								height: 250,
									buttons: {
        								"OK": function() {
											var obj = {Name: $("#teName").val(),
														Url: $("#teUrl").val(),
														CheckFuncName: $("#teChkType").val(),
														Emails: $("#teEmails").val()}
										   saveObject(obj, true);
        								   $( this ).dialog( "close" );
        								},
      		  							"Cancel": function() {$( this ).dialog( "close" );}
      								}});
	$( "#delDialog" ).dialog({autoOpen: false,
	 							modal: true,
								title: "Delete",
								width: 400,
								height: 200});
	$( "#errDialog" ).dialog({autoOpen: false,
	 							modal: true,
								title: "Error",
								width: 400,
								height: 200, 
								buttons: {"Close": function() {$( this ).dialog( "close" );}}});
	$( "#btAdd").button().click(function(){
		$("#teName").attr("disabled", false);
		$("#teName").val("");
		$("#teUrl").val("");
		$("#teChkType").val("");
		$("#teEmails").val("");
		$("#addDialog" ).dialog({title: "Add new record", 
								buttons: {
        								"OK": function() {
											var obj = {Name: $("#teName").val(),
														Url: $("#teUrl").val(),
														CheckFuncName: $("#teChkType").val(),
														Emails: $("#teEmails").val()}
										   saveObject(obj, $( this ), true);
        								},
      		  							"Cancel": function() {$( this ).dialog( "close" );}
      								}});
		$( "#addDialog" ).dialog("open");
	});
	refreshTable();
	$("body").show();
});

function refreshTable(){
	$(".row.content").remove();
	$.getJSON("/data", function(data){
		if(!data){
			return;
		}
		var rows = [];
		for(var i=0; i<data.length; i++){
			var obj  = data[i];
			rows.push(createRow(obj));
		}
		$("#data").append(rows);
	});
}
function createRow(obj){
	var row = $("<div>").addClass("row").addClass("content");
	row.data("obj", obj);
	var btEdit = $("<div>").text("Edit");
	btEdit.data("row", row);
	btEdit.button().click(function(){
		var objData = $(this).data("row").data("obj");
		$("#teName").attr("disabled", true);
		$("#teName").val(objData .Name);
		$("#teUrl").val(objData .Url);
		$("#teChkType").val(objData .CheckFuncName);
		$("#teEmails").val(objData .Emails);
		$("#addDialog" ).dialog({title: "Edit record", 
									buttons: {
        								"OK": function() {
											var obj = {Name: $("#teName").val(),
														Url: $("#teUrl").val(),
														CheckFuncName: $("#teChkType").val(),
														Emails: $("#teEmails").val()}
										   saveObject(obj, $( this ), false);
        								},
      		  							"Cancel": function() {$( this ).dialog( "close" );}
      								}});
		$("#addDialog").dialog("open");				
	});
	var btDelete = $("<div>").text("Delete");
	btDelete.data("row", row);
	btDelete.button().click(function(){
		var crow = $(this).data("row");
		$("#delDialogText").text("Delete record with name '"+crow.data("obj").Name+"'?");
		$("#delDialog").dialog({buttons:[
			{text: "OK",
			click: function(){
				var obj = crow.data("obj");
				$.post("/del", obj, function(data){
					if(!handleError(data)){
						crow.remove();
						$("#delDialog").dialog("close");		
					}
				});	
				}
			},
			{text: "Cancel", click: function() {$(this).dialog("close")}}
			]});
			$("#delDialog").dialog("open");				
	});
	row.append($("<div>").addClass("cell").addClass("w10").text(obj.Name));
	var url = obj.Url;
	row.append($("<div>").addClass("cell").addClass("w40").append($("<a>").attr("href",url).text(url)));
	row.append($("<div>").addClass("cell").addClass("w10").text(obj.CheckFuncName));
	row.append($("<div>").addClass("cell").addClass("w20").text(obj.Emails));
	row.append($("<div>").addClass("cell").addClass("w5").append(btEdit));
	row.append($("<div>").addClass("cell").addClass("w5").append(btDelete));
	//$("#data").append(row);
	return row
}
function updateRow(obj){
	var crow;
	var rows = $("div .row.content");
	for (var i=0; i<rows.length;i++){
		var row = $(rows[i]);
		if(obj.Name === row.data("obj").Name){
			crow = row;
			break;
		}
	}
	var url = obj.Url;
	$(crow.find("div:eq(1)>a")[0]).attr("href",url).text(url);
	$(crow.find("div:eq(2)")[0]).text(obj.CheckFuncName);
	$(crow.find("div:eq(3)")[0]).text(obj.Emails);
	crow.data("obj", obj);
}

function saveObject(obj, dialog, add){
	var addr = "/save";
	if(add){
		addr = "/add"
	}
	$.post(addr, obj, function(data){
		if(!handleError(data)){
			if(add){
				$("#data").append(createRow(obj));
			}else{
				updateRow(obj);
			}
			dialog.dialog("close");
		}
	} );
}

function handleError(data){
	var error = data != "";
	if(error){
		$("#errDialogText").text(data);
		$("#errDialog").dialog("open");
	}
	return error;
}