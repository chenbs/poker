--Poker=class()
--shuffle洗牌、deal发牌

p1={seat=1,hand={}}
p2={seat=2,hand={}}
p3={seat=3,hand={}}
p4={seat=4,hand={}}
handstr1=""
handstr2=""
handstr3=""
handstr4=""
function ctor(specard,rank)

    specard=specard or 2
    local rank_=rank
    poker={}
    restnum=108

    for i=1, restnum do
        poker[i]=i
    end
    xipai()
    fapai()
    paixu(specard)
    if #rank_==4 then
        rank={}
        for v in string.gmatch(rank_,"%d") do
            if tonumber(v)==1 then
                rank[#rank+1]=p1
            elseif tonumber(v)==2 then
                rank[#rank+1]=p2
            elseif tonumber(v)==3 then
                rank[#rank+1]=p3
            else
                rank[#rank+1]=p4
            end
        end
        for i,v in pairs(rank) do 
            print(rank[i].hand[1]) 
        end
        gong(specard,rank)
        handle()
    else
        handle()
    end
    --return hgs
end


function xipai()
    math.randomseed(os.time())
    for i=1,restnum do
        local t1=math.random(1,restnum)
        local t2=math.random(1,restnum)
        if t1~=t2 then
            poker[t1],poker[t2]=poker[t2],poker[t1]
        end
    end
end
function fapai()
    for i=1,restnum do
        if i%4==0 then
            p4.hand[#p4.hand+1]=poker[i]
        elseif i%4==3 then
            p3.hand[#p3.hand+1]=poker[i]
        elseif i%4==2 then
            p2.hand[#p2.hand+1]=poker[i]
        else
            p1.hand[#p1.hand+1]=poker[i]
        end
    end
end
function paixu(specard)
    p1.hand=sort(specard,p1.hand)
    p2.hand=sort(specard,p2.hand)
    p3.hand=sort(specard,p3.hand)
    p4.hand=sort(specard,p4.hand)
end

function handle()
    for i=1,#p1.hand do
        handstr1=handstr1..tostring(p1.hand[i]).." "
    end
    for i=1,#p2.hand do
        handstr2=handstr2..tostring(p2.hand[i]).." "
    end
    for i=1,#p3.hand do
        handstr3=handstr3..tostring(p3.hand[i]).." "
    end
    for i=1,#p4.hand do
        handstr4=handstr4..tostring(p4.hand[i]).." "
    end
end

function sort(s,t)
    s=s or 2
    for i=1,#t-1 do
        local max=i
        for j=i+1,#t  do
            if t[max]>104 then
                if (t[max]==105 and t[j]>105) or
                    (t[max]==106 and t[j]==108) or
                    (t[max]==107 and (t[j]==106 or t[j]==108))then
                    max=j
                end
            elseif t[max]%13==s-1 then
                if t[j]>104 then
                    max=j
                elseif t[j]%13==s-1 and math.floor(t[j]/13)%4+1<math.floor(t[max]/13)%4+1 then
                    max=j
                end
            else
                if t[j]>104 or t[j]%13==s-1 then
                    max=j
                elseif t[j]%13==0 and t[max]%13~=0 then
                    max=j 
                elseif t[max]%13==t[j]%13 and math.floor(t[j]/13)%4+1<math.floor(t[max]/13)%4+1 then
                    max=j

                elseif t[max]%13<t[j]%13 and t[max]%13~=0 then
                    max=j
                end
            end
        end
        if max~=i then
            
            -- print(t[max])
            t[max],t[i]=t[i],t[max]
        end
    end
    return t
end

function gong(specard,rank)
    gp={}
    --双下
    if (rank[1].seat==rank[2].seat+2) or (rank[1].seat==rank[2].seat-2) then
        --抗贡,上游先出牌
        if (rank[3].hand[1]==108 and rank[4].hand[1]==106) or
            (rank[4].hand[1]==108 and rank[3].hand[1]==106) or
            (rank[3].hand[1]==108 and rank[3].hand[2]==106) or
            (rank[4].hand[1]==108 and rank[4].hand[2]==106) then
            print("抗贡")
            turn=rank[1].seat
        --双贡
        else
            gp[1]=table.remove(rank[3].hand,xuangongpai(specard, rank[3].hand))
            gp[2]=table.remove(rank[4].hand,xuangongpai(specard, rank[4].hand))
            
            gp1=sort(specard,gp)
            
            --上游拿的rank3的贡牌
            if gp1[1]==gp[1] then
                rank[1].hand[#rank[1].hand+1]=gp[1]
                rank[2].hand[#rank[2].hand+1]=gp[2]
                hg={rank[3].seat,rank[4].seat}
                hgs="["..tostring(hg[1])..","..tostring(hg[2]).."]"
                print("huanggong:",rank[3].seat,rank[4].seat)
                print(rank[3].seat.."进贡给"..rank[1].seat..":"..gp[1].."\n",rank[4].seat.."进贡给"..rank[2].seat..":"..gp[2])
                print(#rank[1].hand,#rank[2].hand,#rank[3].hand,#rank[4].hand)
            --上游拿的rank4的贡牌
            else
                rank[1].hand[#rank[1].hand+1]=gp[2]
                rank[2].hand[#rank[2].hand+1]=gp[1]
                hg={rank[4].seat,rank[3].seat}
                hgs="["..tostring(hg[1])..","..tostring(hg[2]).."]"
                print("huanggong:",rank[4].seat,rank[3].seat)
                print(rank[4].seat.."进贡给"..rank[1].seat..":"..gp[2]"\n"..rank[3].seat.."进贡给"..rank[2].seat..":"..gp[1])
                print(#rank[1].hand,#rank[2].hand,#rank[3].hand,#rank[4].hand)
            end
            rank[1].hand=sort(specard,rank[1].hand)
            rank[2].hand=sort(specard,rank[2].hand)
        end
    else
        --抗贡,上游先出牌
        if (rank[4].hand[1]==108 and rank[4].hand[2]==106) then
            print("抗贡")
            turn=rank[1].seat
        else
            gp[1]=table.remove(rank[4].hand,xuangongpai(specard, rank[4].hand))
            rank[1].hand[#rank[1].hand+1]=gp[1]
            hg={rank[4].seat}
            hgs="["..tostring(hg[1]).."]"
            print(rank[4].seat.."进贡给"..rank[1].seat..":"..gp[1])
            print(#rank[1].hand,#rank[2].hand,#rank[3].hand,#rank[4].hand)
            rank[1].hand=sort(specard,rank[1].hand)
        end
    end
end
function xuangongpai(specard,hand)
    local hand = hand 
    if (math.ceil(hand[1]/13)%4==1 and hand[1]%13==specard-1) then
        if (math.ceil(hand[2]/13)%4==1 and hand[2]%13==specard-1) then
            return 3
        else
            return 2
        end
    else
        return 1
    end
end


