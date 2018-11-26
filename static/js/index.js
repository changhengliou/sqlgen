(function () {
  var searchKey = document.getElementById("search-key"),
    searchBtn = document.getElementById("search-btn");
  window.sessionStorage.setItem("currPage", 1);
  document.forms[0].addEventListener("submit", function (e) {
    e.preventDefault();
  });
  searchKey.addEventListener("focusout", function (e) {
    setTimeout(100, function () {
      var ul = document.getElementById("input-suggest");
      ul.classList.remove("show");
      removeChildren(ul);
    });
  });
  setTimeout(function () {
    searchKey.addEventListener("paste", function (e) {
      onSearchKeyChange(e.clipboardData.getData('text/plain'));
    });
  }, 100);
  searchKey.addEventListener("keypress", function (e) {
    if (e.key === "Enter") {
      e.preventDefault();
      searchBtn.click();
    } else {
      onSearchKeyChange(e.target.value + e.key);
    }
  });

  searchBtn.addEventListener("click", function (e) {
    e.preventDefault();
    if (window.sessionStorage.getItem("lock") === "true") return;
    var keyword = document.getElementById("search-key").value,
      currPage = window.sessionStorage.getItem("currPage"),
      ul = document.getElementById("article-list");
    if (keyword.trim().length === 0) {
      togglePopover("Please at least type some words =)");
      return;
    }
    // loading
    window.sessionStorage.setItem("lock", "true");
    togglePopover();
    ul.classList.remove("error");
    ul.setAttribute("data-text", "Loading...");
    removeChildren(ul);
    this.disabled = "disabled";
    renderPagination();

    var request = new XMLHttpRequest(),
      searchBtn = this;
    request.open("GET", "/api/v1/search?q=" + keyword + "&p=" + currPage);
    request.onreadystatechange = function () {
      if (this.readyState === 4) {
        // done
        window.sessionStorage.setItem("lock", "false");
        removeChildren(ul);
        ul.setAttribute("data-text", "No Result.");
        searchBtn.disabled = "";

        var response = JSON.parse(this.response);
        if (this.status >= 200 && this.status < 300) {
          var data = response.data,
            count = response.count;
          renderArticle(data, ul);
          renderPagination(count, currPage);
        } else {
          ul.setAttribute("data-text", response.error);
          ul.classList.add("error");
        }
      }
    };
    request.send();
  });
  var removeChildren = function (el) {
    while (el.hasChildNodes()) {
      el.removeChild(el.firstChild);
    }
  };
  var togglePopover = function (text) {
    var popover = document.querySelector("div.popover.show");
    if (popover) {
      if (text === undefined || text.length === 0) {
        popover.parentElement.removeChild(popover);
      } else {
        popover.children[1].innerHTML = text;
      }
    } else {
      if (text === undefined || text.length === 0) {
        return;
      } else {
        var div = document.createElement("div"),
          arrow = document.createElement("div"),
          body = document.createElement("div"),
          form = document.getElementById("search-form");
        div.className = "popover show bs-popover-bottom";
        div.style.willChange = "transform";
        div.style.position = "absolute";
        div.style.top = "36px";
        div.style.right = "10px";
        arrow.className = "arrow";
        arrow.style.top = "-8px";
        body.className = "popover-body";
        body.innerText = text;
        body.style.color = "red";
        body.style.padding = "8px";
        body.style.fontSize = "14px";
        div.appendChild(arrow);
        div.appendChild(body);
        form.appendChild(div);
      }
    }
  };
  var formatText = function (text) {
    var regex = /[\t \n]/g,
      result,
      firstIndex = -1,
      lastIndex = -1;
    while ((result = regex.exec(text)) !== null) {
      if (firstIndex < 0) {
        firstIndex = result.index;
      }
      lastIndex = result.index;
    }
    if (firstIndex === -1)
      return "... " + text + " ...";
    if (firstIndex + 1 === text.length)
      return "... " + text.substr(0, firstIndex) + " ...";
    if (firstIndex === lastIndex)
      return "... " + text.substr(firstIndex + 1) + " ...";
    return "... " + text.substr(firstIndex + 1, lastIndex - firstIndex - 1) + " ...";
  };
  var renderPagination = function (count, currPage) {
    var pagination = document.getElementById("pagination");
    removeChildren(pagination);
    if (!arguments.length) return;
    currPage = parseInt(currPage);
    var totalPages = count / 10,
      renderPagesMin = currPage - 5,
      renderPagesMax = currPage + 5;

    totalPages = count % 10 === 0 ? totalPages : totalPages + 1;
    renderPagesMin = renderPagesMin < 1 ? 1 : renderPagesMin;
    renderPagesMax = renderPagesMax > totalPages ? totalPages : renderPagesMax;

    var arr = ["Prev"];
    for (var i = renderPagesMin; i <= renderPagesMax; i++)
      arr.push(i);
    arr = arr.concat("Next");

    for (var i = 0; i < arr.length; i++) {
      var li = document.createElement("li"),
        anchor = document.createElement("a");
      if (arr[i] === currPage) {
        li.classList.add("active");
      }
      li.setAttribute("data-page", arr[i]);
      li.classList.add("page-item");
      li.addEventListener('click', onPageChange);
      anchor.innerText = arr[i];
      anchor.classList.add("page-link");
      li.appendChild(anchor);
      pagination.appendChild(li);
    }
    if (currPage === 1) {
      pagination.firstElementChild.classList.add("disabled");
    }
    if (currPage === totalPages) {
      pagination.lastElementChild.classList.add("disabled");
    }
  };
  var renderArticle = function (data, ul) {
    for (var i = 0; i < data.length; i++) {
      var li = document.createElement("li"),
        div = document.createElement("div"),
        h4 = document.createElement("h4"),
        p = document.createElement("p");
      li.setAttribute("data-id", data[i].id);
      li.className = "article-preview";
      h4.innerText = data[i].title;
      p.innerText = formatText(data[i].text);
      div.appendChild(h4);
      div.appendChild(p);
      li.appendChild(div);
      ul.appendChild(li);
    }
  };
  var onPageChange = function (e) {
    var elem = e.target, currPage = "1";
    if (e.target.tagName === 'A') {
      elem = elem.parentElement;
    }
    currPage = elem.getAttribute("data-page");
    if (currPage === "Prev") {
      if (elem.classList.contains("disabled")) return;
      currPage = parseInt(sessionStorage.getItem("currPage")) - 1;
    }
    if (currPage === "Next") {
      if (elem.classList.contains("disabled")) return;
      currPage = parseInt(sessionStorage.getItem("currPage")) + 1;
    }
    window.sessionStorage.setItem("currPage", currPage);
    document.getElementById("search-btn").click();
  };
  var onSearchKeyChange = function (v) {
    var ul = document.getElementById("input-suggest");
    removeChildren(ul);
    ul.classList.remove("show");

    if (v.length < 3) return;

    var request = new XMLHttpRequest();
    request.open("GET", "/api/v1/suggest?k=" + v);
    request.onreadystatechange = function () {
      if (this.readyState === 4) {
        var response = JSON.parse(this.responseText);
        if (this.status >= 200 && this.status < 300) {
          if (!Array.isArray(response) || response.length === 0) return;
          ul.classList.add("show");
          for (var i = 0; i < response.length; i++) {
            var li = document.createElement("li"),
              a = document.createElement("a");
            li.classList.add("dropdown-list");
            a.href = "/article/" + response[i].id;
            a.innerText = response[i].title;
            li.appendChild(a);
            ul.appendChild(li);
          }
        } else {
          var err = response.error;
          console.log(err);
        }
      }
    };
    request.send();
  };
})();
