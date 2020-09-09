from django.shortcuts import render, get_object_or_404
from django.http import HttpResponse, HttpResponseRedirect
from .models import Question
from .models import Group, Player, MatchID
from django.template import loader
from django.urls import reverse
from django.views import generic

# Create your views here.
def index2(request):
    latest_question_list = Question.objects.order_by('-pub_date')[:5]
    template = loader.get_template('polls/index.html')
    context = {
        'latest_question_list': latest_question_list,
    }
    return HttpResponse(template.render(context, request))
    #output = ', '.join([q.question_text for q in latest_question_list])
    #return HttpResponse(output)

def index(request):
    latest_match_date = MatchID.objects.order_by("-match_id")[0]
    latest_group_list = Group.objects.filter(match_id=latest_match_date)
    latest_player_list = Player.objects.filter(match_id=latest_match_date)
    latest_match_id_list = MatchID.objects.order_by("-match_id")[:5]

    return render(request, 'polls/index.html', {'latest_group_list':latest_group_list, 'latest_player_list':latest_player_list, 'latest_match_id_list':latest_match_id_list}) 

# 比赛结果详情
def detail(request, match_date):
    latest_group_list = Group.objects.filter(match_id=match_date)
    latest_player_list = Player.objects.filter(match_id=match_date)
    latest_match_id_list = MatchID.objects.order_by("-match_id")[:5]

    #latest_group_list = Group.objects.order_by("-match_id")[:21]
    return render(request, 'polls/detail.html', {'latest_group_list':latest_group_list, 'latest_player_list':latest_player_list, 'latest_match_id_list':latest_match_id_list})

def detail2(request, question_id):
    question = get_object_or_404(Question, pk=question_id)
    return render(request, 'polls/detail.html', {'question': question})

def results(request, question_id):
    question = get_object_or_404(Question, pk=question_id)
    return render(request, 'polls/results.html', {'question': question})

def vote(request, question_id):
    question = get_object_or_404(Question, pk=question_id)
    try:
        selected_choice = question.choice_set.get(pk=request.POST['choice'])
    except (KeyError, Choice.DoesNotExist):
        # Redisplay the question voting form.
        return render(request, 'polls/detail.html', {
            'question': question,
            'error_message': "You didn't select a choice.",
        })
    else:
        selected_choice.votes += 1
        selected_choice.save()
        # Always return an HttpResponseRedirect after successfully dealing
        # with POST data. This prevents data from being posted twice if a
        # user hits the Back button.
        return HttpResponseRedirect(reverse('polls:results', args=(question.id,)))
