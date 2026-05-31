// AGGridPanel init — docs/reference/ag-grid-ssrm-contract.md
// Parameterised by data-grid-endpoint and data-grid-target on the container.
// AG Grid Enterprise 35.x; rowModelType:'serverSide'; theme:'legacy'.
// No license key — unlicensed watermark is the accepted demo tradeoff.
(function () {
  'use strict';
  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-grid-endpoint]').forEach(function (el) {
      var endpoint = el.getAttribute('data-grid-endpoint');
      var target = el.getAttribute('data-grid-target');
      var rowClick = el.getAttribute('data-grid-rowclick') || 'detail';
      // Overlay text is the same for all roles; the "New Project" create button is
      // a server-rendered element on the page (present/absent server-decided, per AC7).
      var noRowsTemplate = '<span class="ag-overlay-no-rows-center">No projects yet — create one to get started</span>';

      var columnDefs = [
        { field: 'code',                     headerName: 'Code',              filter: true, sortable: true },
        { field: 'name',                     headerName: 'Name',              filter: true, sortable: true },
        { field: 'status',                   headerName: 'Status',            filter: 'agSetColumnFilter', filterParams: { values: ['Active','OnHold','Closed'] }, sortable: true },
        { field: 'compliance_score',         headerName: 'Compliance Score',  filter: 'agNumberColumnFilter', sortable: true },
        { field: 'start_date',               headerName: 'Start Date',        filter: 'agDateColumnFilter',  sortable: true },
        { field: 'target_completion_date',   headerName: 'Target Completion', filter: 'agDateColumnFilter',  sortable: true }
      ];

      var datasource = {
        getRows: function (params) {
          fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest' },
            body: JSON.stringify(params.request)
          })
          .then(function (r) { return r.ok ? r.json() : Promise.reject(r.status); })
          .then(function (d) { params.success({ rowData: d.rows, rowCount: d.lastRow }); })
          .catch(function ()  { params.fail(); });
        }
      };

      agGrid.createGrid(el, {
        theme: 'legacy',
        columnDefs: columnDefs,
        defaultColDef: { filter: true, sortable: true },
        rowModelType: 'serverSide',
        serverSideDatasource: datasource,
        overlayNoRowsTemplate: noRowsTemplate,
        onRowClicked: function (e) {
          if (!e.data || !e.data.id) {
            return;
          }
          if (rowClick === 'navigate') {
            window.location.href = '/projects/' + encodeURIComponent(e.data.id);
            return;
          }
          if (target) {
            htmx.ajax('GET', '/projects/' + e.data.id, { target: target, swap: 'innerHTML' });
          }
        }
      });
    });
  });
}());
